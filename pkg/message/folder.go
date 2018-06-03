// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package message

type Folder struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

func (this Folder) IsPayload() {}
