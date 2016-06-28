/* Copyright (C) Couchbase, Inc 2016 - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 */

package man

import (
	"fmt"
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
	return fmt.Sprintf("Error executing %s command on %s, `%s`", e.exe,
		e.path, e.err.Error())
}
