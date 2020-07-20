// Copyright (c) 2020. Temple3x (temple3x@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/zaibyte/pkg/config"
	"github.com/zaibyte/pkg/uid"
	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/xlog"
	"github.com/zaibyte/pkg/xnet"

	"github.com/julienschmidt/httprouter"
)

// ServerConfig is the config of Server.
type ServerConfig struct {
	BoxID   uint32
	Address string

	Encrypted         bool
	CertFile, KeyFile string

	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

const (
	defaultIdleTimeout       = 75 * time.Second
	defaultReadHeaderTimeout = 3 * time.Second
)

// Server implements methods to build & run a HTTP server.
type Server struct {
	cfg *ServerConfig

	router *httprouter.Router
	srv    *http.Server

	onFinish []func() error // run these functions before exit
}

func parseConfig(cfg *ServerConfig) {
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		cfg.Encrypted = false
	}

	config.Adjust(&cfg.IdleTimeout, defaultIdleTimeout)
	config.Adjust(&cfg.ReadHeaderTimeout, defaultReadHeaderTimeout)
}

// NewServer creates a Server.
//
// Warn: Be sure you have run InitGlobalLogger before call it.
func NewServer(cfg *ServerConfig) (s *Server) {

	parseConfig(cfg)

	s = &Server{
		cfg: cfg,
	}

	s.addDefaultHandler()
	s.addDefaultOnFinish()

	s.srv = &http.Server{
		Addr:     cfg.Address,
		ErrorLog: log.New(xlog.GetLogger(), "", 0),
		Handler:  s.router,

		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return
}

// HandlerFunc wraps http.HandlerFunc and returns written & status.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, p httprouter.Params) (written, status int)

// AddHandler helps to add handler to Server.
func (s *Server) AddHandler(method, path string, handler HandlerFunc, limit int64) {
	if limit > 0 {
		l := newReqLimit(limit)
		handler = l.withLimit(handler)
	}
	s.router.Handle(method, path, s.must(handler))
}

// AddOnFinish adds function that will run before exit.
func (s *Server) AddOnFinish(f func() error) {
	s.onFinish = append(s.onFinish, f)
}

func (s *Server) addDefaultOnFinish() {
	s.AddOnFinish(xlog.Sync)
}

// Run starts the Server and implements graceful shutdown.
func (s *Server) Run() {

	go func() {
		if s.cfg.Encrypted && s.cfg.CertFile != "" && s.cfg.KeyFile != "" {
			if err := s.srv.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := s.srv.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
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

	for _, f := range s.onFinish {
		f()
	}

	switch sig {
	case syscall.SIGTERM:
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

// must adds the headers which zai must have and check request body.
func (s *Server) must(next HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		reqID := r.Header.Get(ReqIDHeader)
		if reqID == "" {
			reqID = uid.MakeReqID(s.cfg.BoxID)
		}
		w.Header().Set(ReqIDHeader, reqID)

		if !s.cfg.Encrypted {
			checksum, err := strconv.Atoi(r.Header.Get(ChecksumHeader))
			if err != nil {
				ReplyError(w, ErrHeaderCheckFailedMsg, http.StatusBadRequest)
				return
			}

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ReplyCode(w, http.StatusInternalServerError)
				return
			}

			if checksum != int(xnet.Checksum(b)) {
				ReplyError(w, ErrHeaderCheckFailedMsg, http.StatusBadRequest)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewReader(b))
		}

		next(w, r, p)
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

// addDefaultHandler add default handler.
func (s *Server) addDefaultHandler() {
	if s.router == nil {
		s.router = httprouter.New()
	}

	s.AddHandler(http.MethodPut, "/v1/debug-log/:cmd", s.debug, 1)
	s.AddHandler(http.MethodGet, "/v1/code-version", s.version, 1)
}

func (s *Server) debug(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) (written, status int) {

	reqID := w.Header().Get(ReqIDHeader)

	cmd := p.ByName("cmd")
	switch cmd {
	case "on":
		xlog.SetLevel("debug")
		xlog.DebugID(reqID, "debug on")
	default:
		xlog.SetLevel("info")
		xlog.InfoID(reqID, "debug off")
	}

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

// Reply replies HTTP request, return the written bytes length & status code,
// we need the status code for access log.
//
// Usage:
// As return function in http Handler.
//
// Warn:
// Be sure you have called xlog.InitGlobalLogger.
// If any wrong in the write resp process, it would be written into the log.

// ReplyCode replies to the request with the empty message and HTTP code.
func ReplyCode(w http.ResponseWriter, statusCode int) (written, status int) {

	return ReplyJson(w, nil, statusCode)
}

// ReplyError replies to the request with the specified error message and HTTP code.
func ReplyError(w http.ResponseWriter, msg string, statusCode int) (written, status int) {

	if msg == "" {
		msg = http.StatusText(statusCode)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	written, err := fmt.Fprintln(w, msg)
	if err != nil {
		xlog.ErrorID(w.Header().Get(ReqIDHeader), makeReplyErrMsg(err))
	}
	return written, statusCode
}

// ReplyJson replies to the request with specified ret(in JSON) and HTTP code.
func ReplyJson(w http.ResponseWriter, ret interface{}, statusCode int) (written, status int) {

	var msg []byte
	if ret != nil {
		msg, _ = json.Marshal(ret)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
	w.WriteHeader(statusCode)
	written, err := w.Write(msg)
	if err != nil {
		xlog.ErrorID(w.Header().Get(ReqIDHeader), makeReplyErrMsg(err))
	}
	return written, statusCode
}

// ReplyBin replies to the request with specified ret(in Binary) and length.
func ReplyBin(w http.ResponseWriter, ret io.Reader, length int64) (written, status int) {

	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	n, err := io.CopyN(w, ret, length)
	if err != nil {
		xlog.ErrorID(w.Header().Get(ReqIDHeader), makeReplyErrMsg(err))
	}
	return int(n), http.StatusOK
}

func makeReplyErrMsg(err error) string {
	return fmt.Sprintf("write resp failed: %s", err.Error())
}

// FillPath fills the julienschmidt/httprouter style path.
func FillPath(path string, kv map[string]string) string {
	if kv == nil {
		return path
	}

	for k, v := range kv {
		path = strings.Replace(path, ":"+k, v, 1)
	}
	return path
}
