// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package message

import (
	"encoding/json"
	"github.com/go-errors/errors"
)

type Type string

const (
	FolderContents Type = "FolderContents"
)

type Message struct {
	Type    Type    `json:"type"`
	Payload Payload `json:"payload"`
}

func (this *Message) UnmarshalJSON(data []byte) error {
	var tmpType struct {
		Type Type `json:"type"`
	}

	err := json.Unmarshal(data, &tmpType)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	this.Type = tmpType.Type

	switch this.Type {
	case FolderContents:
		var tmpPayload struct {
			Payload Folder `json:"payload"`
		}

		err := json.Unmarshal(data, &tmpPayload)
		if err != nil {
			return errors.Wrap(err, 0)
		}

		this.Payload = tmpPayload.Payload
	default:
		return errors.Errorf("unhandled type of message (%s)", this.Type)
	}

	return nil
}
