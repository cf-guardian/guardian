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

package error_test

import (
	"errors"
	"github.com/cf-guardian/guardian/error"
	"strings"
	"testing"
)

const testMessage = "test message"

func TestMessageCapture(t *testing.T) {

	e := error.New(testMessage)

	actual := e.Error()

	if !strings.Contains(actual, testMessage) {
		t.Errorf("%q does not contain %q", actual, testMessage)
	}
}

func TestStackTraceCapture(t *testing.T) {

	const stackPortion = "error_test.TestStackTraceCapture"

	e := error.New(testMessage)

	actual := e.Error()

	if !strings.Contains(actual, stackPortion) {
		t.Errorf("%q does not contain %q", actual, stackPortion)
	}
}

func TestFromError(t *testing.T) {

	cause := errors.New(testMessage)

	e := error.FromError(cause)

	actual := e.Error()

	if !strings.Contains(actual, testMessage) {
		t.Errorf("%q does not contain %q", actual, testMessage)
	}
}
