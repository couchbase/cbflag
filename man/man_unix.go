// +build !windows

/* Copyright (C) Couchbase, Inc 2016 - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 */

package man

import (
	"os"
	"os/exec"
	"path/filepath"
)

func ShowManual(manpath, page string) error {
	path := filepath.Join(manpath, page)

	// Make sure the file exists
	if src, err := os.Stat(path); err != nil {
		return ManPathError{path}
	} else if src.IsDir() {
		return ManPathError{path}
	}

	mcmd := exec.Command("man", path)
	mcmd.Stdout = os.Stdout

	if err := mcmd.Start(); err != nil {
		return ManExecutionError{path, "man", err}
	}

	mcmd.Wait()

	return nil
}
