// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package debouncer

import (
	"time"
)

type ticker struct {
	stopChan chan struct{}
	channel  chan time.Time
	duration time.Duration
	started  bool
}

func newTicker(duration time.Duration) *ticker {
	return &ticker{
		stopChan: make(chan struct{}),
		channel:  make(chan time.Time),
		duration: duration,
		started:  false,
	}
}

func (this *ticker) Start() {
	if this.started {
		return
	}
	this.started = true

	go func() {
		ticker := time.NewTicker(this.duration)

		for {
			select {
			case <-this.stopChan:
				ticker.Stop()
				return
			case tickTime := <-ticker.C:
				this.channel <- tickTime
			}
		}
	}()
}

func (this *ticker) Stop() {
	if !this.started {
		return
	}

	this.started = false
	this.stopChan <- struct{}{}
}

func (this *ticker) Ticks() <-chan time.Time {
	return this.channel
}
