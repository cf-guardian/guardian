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

package gerror_test

import (
	"github.com/cf-guardian/guardian/gerror"
	"os"
)

// Define a suitable tag type and some values.
type ErrorId int

const (
	ErrExample ErrorId = iota
	ErrInvalidPort
	ErrInvalidPath
)

func Example() error {
	gerr := SomeFunc()
	if gerr != nil {
		// Use the tag to check for a specific error.
		if gerr.EqualTag(ErrInvalidPath) {
			// Act on this error ...
		}
		return gerr
	}
	// ...
	return nil
}

func SomeFunc() gerror.Gerror {
	_, err := os.Open("/some/path")
	if err != nil {
		return gerror.NewFromError(ErrInvalidPath, err)
	}
	return nil
}
