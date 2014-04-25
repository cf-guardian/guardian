/*
Copyright 2014 GoPivotal (UK) Limited.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package gerror provides an improved error type which captures the stack trace
at construction time.
*/
package gerror

import (
	"fmt"
	"runtime"
)

const stackSize = 4096

// Returns an error containing the given message.
func New(message string) *err {
	var stack [stackSize]byte

	runtime.Stack(stack[:], false)

	return &err{message, stack[:]}
}

func Newf(format string, insert ...interface{}) *err {
	return New(fmt.Sprintf(format, insert...))
}

func FromError(cause error) error {
	if cause != nil {
		var stack [stackSize]byte

		n := runtime.Stack(stack[:], false)

		return &err{"Error caused by: " + cause.Error(), stack[:n]}
	} else {
		return nil
	}
}

type err struct {
	message    string
	stackTrace []byte
}

func (e *err) Error() string {
	return e.message + "\n" + string(e.stackTrace[:])
}
