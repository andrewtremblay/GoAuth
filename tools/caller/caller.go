package caller

import (
	"path"
	"runtime"
)

// Return path where this function executed
func Path() string {
	_, filename, _, ok := runtime.Caller(1)
	if ok {

		return path.Dir(filename)
	}

	return ""

}
