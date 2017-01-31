/* Copyright (C) Couchbase, Inc 2016 - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 */

package man

import (
	"fmt"
	"os/exec"
)

type ManPathError struct {
	path string
}

func (e ManPathError) Error() string {
	return fmt.Sprintf("Man file `%s` does not exist", e.path)
}

type ManExecutionError struct {
	path string
	exe  string
	err  error
}

func (e ManExecutionError) Error() string {
	errStr := e.err.Error()
	if execErr, ok := e.err.(*exec.Error); ok {
		errStr = execErr.Err.Error()
	}
	return fmt.Sprintf("Unable to open man page using `%s`, %s", e.exe, errStr)
}
