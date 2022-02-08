// Copyright 2021 Google LLC
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

//go:build windows
// +build windows

package namedpipe

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/Microsoft/go-winio"
)

type Fifo struct {
	path     string
	listener net.Listener
	buf      chan []byte
	done     chan error
}

func New(name string, bufCount, bufSize int) (*Fifo, error) {
	tmpPath := fmt.Sprintf("\\\\.\\pipe\\%s-%d", name, rand.Int63())
	listener, err := winio.ListenPipe(tmpPath, &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        false,
		InputBufferSize:    int32(bufSize),
		OutputBufferSize:   int32(bufSize),
	})
	if err != nil {
		return nil, err
	}
	f := &Fifo{
		path:     tmpPath,
		listener: listener,
		buf:      make(chan []byte, bufCount),
		done:     make(chan error),
	}
	return f, nil
}

func (f *Fifo) Path() string {
	return f.path
}

func (f *Fifo) Write(p []byte) (int, error) {
	f.buf <- p
	return len(p), nil
}

func (f *Fifo) Close() error {
	close(f.buf)
	return <-f.done
}

func (f *Fifo) run() {
	err := f.runInternal()
	f.done <- err
	close(f.done)
}

func (f *Fifo) runInternal() (err error) {
	var pipe net.Conn
	pipe, err = f.listener.Accept()
	if err != nil {
		return err
	}
	defer func() {
		errC := pipe.Close()
		if err == nil {
			err = errC
		}
	}()
	err = f.listener.Close()
	if err != nil {
		return err
	}
	f.listener = nil
	for {
		data, ok := <-f.buf
		if !ok {
			return nil
		}
		_, err = pipe.Write(data)
		if err != nil {
			return err
		}
	}
	return pipe.Close()
}