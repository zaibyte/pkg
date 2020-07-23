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

package zhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zaibyte/pkg/zdigest"

	"github.com/julienschmidt/httprouter"
	"github.com/zaibyte/pkg/config"
	"github.com/zaibyte/pkg/uid"
	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/zlog"
)

// ServerConfig is the config of Server.
type ServerConfig struct {
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
	cfg    *ServerConfig
	router *httprouter.Router
	srv    *http.Server
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

	s.srv = &http.Server{
		Addr:     cfg.Address,
		ErrorLog: log.New(zlog.GetLogger(), "", 0),
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

// Start starts the Server.
func (s *Server) Start() {

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
}

// Close closes Server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// must adds the headers which zai must have and check request body.
func (s *Server) must(next HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		reqID := r.Header.Get(ReqIDHeader)
		if reqID == "" {
			reqID = strconv.FormatUint(uid.MakeReqID(), 10)
		}
		w.Header().Set(ReqIDHeader, reqID)

		if !s.cfg.Encrypted {

			incoming, err := strconv.Atoi(r.Header.Get(ChecksumHeader))
			if err != nil {
				ReplyError(w, ErrHeaderCheckFailedMsg, http.StatusBadRequest)
				return
			}

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				ReplyCode(w, http.StatusInternalServerError)
				return
			}

			h := crc32.New(zdigest.CrcTbl)
			h.Write([]byte(r.URL.RequestURI()))
			h.Write(b)
			if incoming != int(h.Sum32()) {
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

func (s *Server) debug(w http.ResponseWriter, _ *http.Request,
	p httprouter.Params) (written, status int) {

	reqIDS := w.Header().Get(ReqIDHeader)
	reqID := reqIDStrToInt(reqIDS)

	cmd := p.ByName("cmd")
	switch cmd {
	case "on":
		_ = zlog.SetLevel("debug")
		zlog.DebugID(reqID, "debug on")
	default:
		_ = zlog.SetLevel("info")
		zlog.InfoID(reqID, "debug off")
	}

	return ReplyCode(w, http.StatusOK)
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) (written, status int) {

	return ReplyJson(w, &version.Info{
		Version:   version.ReleaseVersion,
		GitHash:   version.GitHash,
		GitBranch: version.GitBranch,
	}, http.StatusOK, s.cfg.Encrypted)
}

// Reply replies HTTP request, return the written bytes length & status code.
//
// Usage:
// As return function in http Handler.
//
// Warn:
// Be sure you have called zlog.InitGlobalLogger.
// If any wrong in the write resp process, it would be written into the log.

// ReplyCode replies to the request with the empty message and HTTP code.
func ReplyCode(w http.ResponseWriter, statusCode int) (written, status int) {

	return ReplyJson(w, nil, statusCode, true) // Only reply code, no need check resp.body.
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
		zlog.ErrorID(reqIDStrToInt(w.Header().Get(ReqIDHeader)), makeReplyErrMsg(err))
	}
	return written, statusCode
}

// ReplyJson replies to the request with specified ret(in JSON) and HTTP code.
func ReplyJson(w http.ResponseWriter, ret interface{}, statusCode int, encrypted bool) (written, status int) {

	var msg []byte
	if ret != nil {
		msg, _ = json.Marshal(ret)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
	w.WriteHeader(statusCode)

	if !encrypted {
		w.Header().Set(ChecksumHeader, strconv.FormatInt(int64(zdigest.Checksum(msg)), 10))
	}

	written, err := w.Write(msg)
	if err != nil {
		zlog.ErrorID(reqIDStrToInt(w.Header().Get(ReqIDHeader)), makeReplyErrMsg(err))
	}
	return written, statusCode
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
