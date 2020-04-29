/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xhttp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/zaibyte/pkg/config"

	"github.com/julienschmidt/httprouter"
	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/xlog"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// ServerConfig is the config of Server.
type ServerConfig struct {
	AppName string
	Address string

	MaxConcurrentStreams uint32
	MaxReadFrameSize     uint32
	IdleTimeout          time.Duration
}

const (
	defaultAppName = "-"

	defaultMaxConcurrentStreams uint32 = 250
	defaultMaxReadFrameSize     uint32 = 16 * 1024
	defaultIdleTimeout                 = 75 * time.Second
)

// Server implements methods to build & run a HTTP server.
type Server struct {
	aLog *xlog.AccessLogger

	router *httprouter.Router

	srv *http.Server
	h2  *http2.Server

	exits []func() error // run these functions before exit
}

func parseConfig(cfg *ServerConfig) {
	config.Adjust(&cfg.AppName, defaultAppName)
	config.Adjust(&cfg.IdleTimeout, defaultIdleTimeout)
	config.Adjust(&cfg.MaxConcurrentStreams, defaultMaxConcurrentStreams)
	config.Adjust(&cfg.MaxReadFrameSize, defaultMaxReadFrameSize)
}

// NewServer creates a Server.
//
// Warn: Be sure you have run InitGlobalLogger before call it.
func NewServer(cfg *ServerConfig, aLog *xlog.AccessLogger) (s *Server) {

	parseConfig(cfg)

	s = &Server{
		aLog: aLog,
	}

	s.addDefaultHandler()
	s.withDefaultExit()

	s.srv = &http.Server{
		Addr:     cfg.Address,
		ErrorLog: log.New(xlog.GetLogger(), "", 0),
	}

	s.h2 = &http2.Server{
		IdleTimeout:          cfg.IdleTimeout,
		MaxConcurrentStreams: cfg.MaxConcurrentStreams,
		MaxReadFrameSize:     cfg.MaxReadFrameSize,
	}

	return
}

// HandlerFunc wraps http.HandlerFunc and returns written & status for access Log.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, p httprouter.Params) (written, status int)

// AddHandler helps to add handler to Server.
func (s *Server) AddHandler(name, method, path string, handler HandlerFunc, limit int64) {
	if limit > 0 {
		l := newReqLimit(limit)
		handler = l.withLimit(handler)
	}
	s.router.Handle(method, path, s.withLog(handler, name))
}

func (s *Server) AddExit(f func() error) {
	s.exits = append(s.exits, f)
}

// Run starts the Server and implements graceful shutdown.
func (s *Server) Run() {

	s.srv.Handler = s.toH2CHandler()

	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-c
	msg := fmt.Sprintf("got signal to exit, signal %s", sig.String())
	xlog.Info(msg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	s.srv.Shutdown(ctx)

	for _, f := range s.exits {
		f()
	}

	switch sig {
	case syscall.SIGTERM:
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

// toH2CHandler returns http.Handler with h2c server.
func (s *Server) toH2CHandler() http.Handler {
	return h2c.NewHandler(s.router, s.h2)
}

// with accessLog.
// All handler must be with access log.
//
// ps:
// withLog will also add the headers which zai must have.
func (s *Server) withLog(next HandlerFunc, name string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		// start & err for access log
		start := time.Now()

		w.Header().Set(xlog.BoxIDHeader, strconv.FormatInt(xlog.GetBoxID(), 10))

		reqID := r.Header.Get(xlog.ReqIDHeader)
		if reqID == "" {
			reqID = xlog.NextReqID()
		}
		w.Header().Set(xlog.ReqIDHeader, reqID)

		written, status := next(w, r, p)

		s.aLog.Write(name, r, start, reqID, written, status)
	}
}

// reqLimit implements the ability to limit request count at the same time.
type reqLimit struct {
	limit int64
	cnt   int64
}

func newReqLimit(limit int64) *reqLimit {
	return &reqLimit{
		limit: limit,
	}
}

func (l *reqLimit) withLimit(next HandlerFunc) HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) (written, status int) {

		if atomic.AddInt64(&l.cnt, 1) > l.limit {
			atomic.AddInt64(&l.cnt, -1)
			return ReplyCode(w, http.StatusTooManyRequests)
		}
		written, status = next(w, r, p)
		atomic.AddInt64(&l.cnt, -1)
		return
	}
}

// --- Default Handler ---- //
// --- All HTTP Servers in zai will have these APIs ---- //
// Don't forget to add new API to xhttp.Client.
const (
	debugAPIName   = "debug"
	versionAPIName = "version"
	pingAPIName    = "ping" // ping is used for checking server health and get the boxID.
)

// addDefaultHandler add default handler.
func (s *Server) addDefaultHandler() {
	if s.router == nil {
		s.router = httprouter.New()
	}

	s.AddHandler(debugAPIName, http.MethodPut, "/v1/debug-log/:cmd", s.debug, 1)
	s.AddHandler(versionAPIName, http.MethodGet, "/v1/code-version", s.version, 1)
	s.AddHandler(pingAPIName, http.MethodHead, "/v1/ping", s.ping, 1)
}

func (s *Server) debug(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) (written, status int) {

	reqID := w.Header().Get(xlog.ReqIDHeader)

	cmd := p.ByName("cmd")
	switch cmd {
	case "on":
		xlog.DebugOn()
		xlog.DebugWithReqID("debug on", reqID)
	default:
		xlog.DebugOff()
		xlog.InfoWithReqID("debug off", reqID)
	}

	return ReplyCode(w, http.StatusOK)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) (written, status int) {

	return ReplyCode(w, http.StatusOK)
}

func (s *Server) version(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) (written, status int) {

	return ReplyJson(w, &version.Info{
		version.ReleaseVersion,
		version.GitHash,
		version.GitBranch,
	}, http.StatusOK)
}

func (s *Server) withDefaultExit() {
	s.AddExit(s.aLog.Sync)
	s.AddExit(xlog.Sync)
}
