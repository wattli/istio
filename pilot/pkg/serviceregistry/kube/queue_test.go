// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"errors"
	"testing"
	"time"

	"istio.io/istio/pilot/pkg/model"
)

func TestQueue(t *testing.T) {
	q := NewQueue(1 * time.Microsecond)
	stop := make(chan struct{})
	done := make(chan struct{})
	out := 0
	err := true
	add := func(obj interface{}, event model.Event) error {
		t.Logf("adding %d, error: %t", obj.(int), err)
		out += obj.(int)
		if out == 4 {
			close(done)
		}
		if !err {
			return nil
		}
		err = false
		return errors.New("intentional error")
	}
	go q.Run(stop)

	q.Push(Task{Handler: add, Obj: 1})
	q.Push(Task{Handler: add, Obj: 1})
	q.Push(Task{Handler: func(Obj interface{}, event model.Event) error {
		out += Obj.(int)
		if out != 3 {
			t.Errorf("Queue => %d, want %d", out, 3)
		}
		return nil
	}, Obj: 1})

	// wait for all task processed
	<-done
	close(stop)
}

func TestChainedHandler(t *testing.T) {
	q := NewQueue(1 * time.Microsecond)
	stop := make(chan struct{})
	done := make(chan struct{})
	out := 0
	f := func(i int) Handler {
		return func(obj interface{}, event model.Event) error {
			out += i
			return nil
		}
	}
	handler := ChainHandler{
		Funcs: []Handler{f(1), f(2)},
	}
	go q.Run(stop)

	q.Push(Task{Handler: handler.Apply, Obj: 0})
	q.Push(Task{Handler: func(obj interface{}, Event model.Event) error {
		if out != 3 {
			t.Errorf("ChainedHandler => %d, want %d", out, 3)
		}
		close(done)
		return nil
	}, Obj: 0})

	<-done
	close(stop)
}
