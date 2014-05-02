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

func ExampleNew() error {
	return gerror.New(ErrExample, "Example error message")
}

func ExampleNewf(portNum int) error {
	return gerror.Newf(ErrInvalidPort, "Invalid port: %d", portNum)
}

func ExampleNewFromError(filePath string) (file *os.File, err error) {
	file, err = os.Open(filePath)
	if err != nil {
		return file, gerror.NewFromError(ErrInvalidPath, err)
	}
	return file, nil
}
