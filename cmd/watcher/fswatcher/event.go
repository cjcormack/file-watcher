// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package fswatcher

type eventType int

const (
	Create eventType = iota
)

type Event struct {
	Folder *Folder
}
