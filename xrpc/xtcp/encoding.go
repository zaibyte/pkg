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

package xtcp

import (
	"bufio"
	"hash"
	"io"
	"net"
	"time"

	"github.com/zaibyte/pkg/xrpc"
)

type timeoutConnReader struct {
	r net.Conn
}

func (r *timeoutConnReader) Read(b []byte) (n int, err error) {
	deadline := time.Now().Add(readDuration)
	if err = r.r.SetReadDeadline(deadline); err != nil {
		return 0, err
	}
	return r.r.Read(b)
}

type decoder struct {
	br   *bufio.Reader
	hash hash.Hash32
}

func newDecoder(conn net.Conn, bufsize int, hash hash.Hash32) *decoder {
	r := &timeoutConnReader{r: conn}
	return &decoder{br: bufio.NewReaderSize(r, bufsize), hash: hash}
}

type message struct {
	header header
	body   xrpc.Byteser
}

func (d *decoder) decode(msg *message) error {

	var hbuf []byte
	_, ok := msg.header.(*reqHeader)
	if ok {
		hbuf = make([]byte, reqHeaderSize)
	} else {
		hbuf = make([]byte, respHeaderSize)
	}
	_, err := io.ReadFull(d.br, hbuf)
	if err != nil {
		operr, ok := err.(net.Error)
		if ok && operr.Timeout() {
			return xrpc.ErrTimeout
		}
		return err
	}
	err = msg.header.decode(hbuf)
	if err != nil {
		return err
	}

	n := msg.header.getBodySize()
	if n == 0 {
		return nil
	}

	if n > xrpc.MaxBytesSizeInPool {
		msg.body = &xrpc.BytesBuffer{
			S: make([]byte, n),
		}
	} else {
		msg.body = xrpc.GetBytes()
	}

	buf := msg.body.Bytes()[:n]
	_, err = readAtLeast(d.br, buf, len(buf), d.hash)
	if err != nil {
		return err
	}
	msg.body.Set(buf)
	return nil
}

// readAtLeast reads from r into buf until it has read at least min bytes.
// It returns the number of bytes copied and an error if fewer bytes were read.
// The error is EOF only if no bytes were read.
// If an EOF happens after reading fewer than min bytes,
// readAtLeast returns ErrUnexpectedEOF.
// If min is greater than the length of buf, ReadAtLeast returns ErrShortBuffer.
// On return, n >= min if and only if err == nil.
// If r returns an error having read at least min bytes, the error is dropped.
func readAtLeast(r io.Reader, buf []byte, min int, hash hash.Hash32) (n int, err error) {
	if len(buf) < min {
		return 0, io.ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		if err == nil && hash != nil {
			hash.Write(buf[n : n+nn])
		}
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

type timeoutConnWriter struct {
	w net.Conn
}

func (w *timeoutConnWriter) Write(b []byte) (n int, err error) {
	deadline := time.Now().Add(writeDuration)
	if err = w.w.SetWriteDeadline(deadline); err != nil {
		return 0, err
	}
	return w.w.Write(b)
}

type encoder struct {
	bw *bufio.Writer
}

func newEncoder(conn net.Conn, bufsize int) *encoder {
	w := &timeoutConnWriter{w: conn}
	return &encoder{bw: bufio.NewWriterSize(w, bufsize)}
}

func (e *encoder) encode(msg *message) error {
	var hbuf []byte
	_, ok := msg.header.(*reqHeader)
	if ok {
		hbuf = make([]byte, reqHeaderSize)
	} else {
		hbuf = make([]byte, respHeaderSize)
	}
	_ = msg.header.encode(hbuf)
	_, err := e.bw.Write(hbuf)
	if err != nil {
		return err
	}

	n := msg.header.getBodySize()
	if n == 0 {
		return nil
	}

	_, err = e.bw.Write(msg.body.Bytes())
	return err
}

func (e *encoder) flush() error {
	return e.bw.Flush()
}
