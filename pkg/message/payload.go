// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package message

type Payload interface {
	// Not needed for anything, but used to tie things to this interface
	IsPayload()
}
