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
