// Copyright 2017 Vlad Didenko. All rights reserved.
// See the included LICENSE.md file for licensing information

package fstests // import "go.didenko.com/fstests"

import (
	"os"
	"path/filepath"
)

func collectFileInfo(dir string) ([]os.FileInfo, error) {

	list := []os.FileInfo{}

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err == nil && path != dir {
			list = append(list, f)
		}
		return err
	})

	return list, err
}
