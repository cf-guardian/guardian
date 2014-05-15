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
Package gerror provides an improved error type which captures an error tag and the stack trace
at construction time.
*/
package gerror

import (
	"fmt"
	"reflect"
	"runtime"
)

const stackSize = 4096

// A Tag represents an error identifier of any type.
type Tag interface{}

// A Gerror is a tagged error with a stack trace embedded in the Error() string.
type Gerror interface {
	// Returns the tag used to create this error.
	Tag() Tag

	// Returns the concrete type of the tag used to create this error.
	TagType() reflect.Type

	// Returns the string form of this error, which includes the tag value, the tag type, the error message, and a stack trace.
	Error() string

	// Test the tag used to create this error for equality with a given tag. Returns `true` if and only if the two are equal.
	EqualTag(Tag) bool
}

// Returns an error containing the given tag and message and the current stack trace.
func New(tag Tag, message string) Gerror {
	var stack [stackSize]byte

	n := runtime.Stack(stack[:], false)

	return &err{tag, reflect.TypeOf(tag), message, stack[:n]}
}

// Returns an error containing the given tag and format string and the current stack trace. The given inserts are applied to the format string to produce an error message.
func Newf(tag Tag, format string, insert ...interface{}) Gerror {
	return New(tag, fmt.Sprintf(format, insert...))
}

// Return an error containing the given tag, the cause of the error, and the current stack trace.
func NewFromError(tag Tag, cause error) Gerror {
	if cause != nil {
		var stack [stackSize]byte

		n := runtime.Stack(stack[:], false)

		return &err{tag, reflect.TypeOf(tag), "Error caused by: " + cause.Error(), stack[:n]}
	} else {
		return nil
	}
}

type err struct {
	tag        Tag
	typ        reflect.Type
	message    string
	stackTrace []byte
}

func (e *err) Error() string {
	return fmt.Sprintf("%v %v", e.tag, e.typ) + ": " + e.message + "\n" + string(e.stackTrace)
}

func (e *err) Tag() Tag {
	return e.tag
}

func (e *err) TagType() reflect.Type {
	return e.typ
}

func (e *err) EqualTag(tag Tag) bool {
	return e.typ == reflect.TypeOf(tag) && e.tag == tag
}
