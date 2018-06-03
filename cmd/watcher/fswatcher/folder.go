// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package fswatcher

import (
	"github.com/cjcormack/file-watcher/pkg/debouncer"
	"github.com/go-errors/errors"
	"io/ioutil"
)

type Folder struct {
	Name      string `json:"name"`
	debouncer *debouncer.Debouncer
}

func (this *Folder) GetCurrentFiles() ([]string, error) {
	fileInfos, err := ioutil.ReadDir(this.Name)
	if err != nil {
		return []string{}, errors.Wrap(err, 0)
	}

	files := make([]string, 0, len(fileInfos))

	for _, fileInfo := range fileInfos {
		// We only care about regular files.
		// TODO Do we want to include SymLinks?
		if fileInfo.Mode().IsRegular() {
			files = append(files, fileInfo.Name())
		}
	}

	return files, nil
}
