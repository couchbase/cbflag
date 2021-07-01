// +build !windows

/* Copyright 2016-Present Couchbase, Inc.
 *
 * Unauthorized copying of this file, via any medium is strictly prohibited
 *
 * Use of this software is governed by the Business Source License included
 * in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
 * in that file, in accordance with the Business Source License, use of this
 * software will be governed by the Apache License, Version 2.0, included in
 * the file licenses/APL2.txt.
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
