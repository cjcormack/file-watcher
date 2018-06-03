// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package debouncer

import (
	"time"
)

type Opts struct {
	Callback func(interface{})
	Duration time.Duration
}

type Debouncer struct {
	c chan interface{}
}

func New(opts Opts) *Debouncer {
	d := &Debouncer{
		c: make(chan interface{}),
	}

	go handleDebounces(d.c, opts)
	return d
}

func (this *Debouncer) Set(i interface{}) {
	this.c <- i
}

func handleDebounces(c chan interface{}, opts Opts) {
	t := newTicker(opts.Duration)

	var lastValue interface{}
	closeNextTick := false

	for {
		select {

		case i := <-c:
			closeNextTick = false

			if !t.started {
				opts.Callback(i)
				lastValue = nil
			} else {
				lastValue = i
			}
			t.Start()
		case <-t.Ticks():
			if closeNextTick {
				t.Stop()
				closeNextTick = false
			} else {
				closeNextTick = true
			}

			if lastValue != nil {
				opts.Callback(lastValue)
			}
			lastValue = nil
		}

	}
}
