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
	"errors"
	"github.com/cf-guardian/guardian/gerror"
	"strings"
	"testing"
)

const testMessage = "test message"

type TestTag int

const (
	BasicError TestTag = iota
	AnotherError
)

func TestMessageCapture(t *testing.T) {

	e := gerror.New(BasicError, testMessage)

	actual := e.Error()

	if !strings.Contains(actual, testMessage) {
		t.Errorf("%q does not contain %q", actual, testMessage)
	}
}

func TestStackTraceCapture(t *testing.T) {

	const stackPortion = "error_test.TestStackTraceCapture"

	e := gerror.New(BasicError, testMessage)

	actual := e.Error()

	if !strings.Contains(actual, stackPortion) {
		t.Errorf("%q does not contain %q", actual, stackPortion)
	}
}

func TestFormatCapture(t *testing.T) {

	e := gerror.Newf(BasicError, "message with %s", "insert")

	actual := e.Error()

	expected := "message with insert"
	if !strings.Contains(actual, expected) {
		t.Errorf("%q does not contain %q", actual, expected)
	}
}

func TestNewFromError(t *testing.T) {

	cause := errors.New(testMessage)

	e := gerror.NewFromError(BasicError, cause)

	actual := e.Error()

	if !strings.Contains(actual, testMessage) {
		t.Errorf("%q does not contain %q", actual, testMessage)
	}
}

func TestNewFromErrorNil(t *testing.T) {

	e := gerror.NewFromError(BasicError, nil)

	if e != nil {
		t.Errorf("e was not nil: %q", e)
	}
}

func TestTagging(t *testing.T) {

	e := gerror.New(AnotherError, "test")

	if id, ok := e.Tag().(TestTag); !ok {
		t.Errorf("Error tag %v has wrong type", id)
	} else if id != AnotherError {
		t.Errorf("Expected %s but got %s", AnotherError, id)
	}

	if e.EqualTag(BasicError) {
		t.Errorf("Error %s has wrong tag", e)
	}

	if !e.EqualTag(AnotherError) {
		t.Errorf("Error %s has wrong tag", e)
	}

	if !strings.HasPrefix(e.Error(), "1 gerror_test.TestTag: ") {
		t.Errorf("Missing error id %s", e.Error())
	}

}

func TestNilTag(t *testing.T) {

	e := gerror.New(nil, testMessage)

	actual := e.Error()

	if !strings.Contains(actual, testMessage) {
		t.Errorf("%q does not contain %q", actual, testMessage)
	}

	if !e.EqualTag(nil) {
		t.Error("Nil tag does not compare equal")
	}

	if e.EqualTag(BasicError) {
		t.Error("Non-nil tag compares equal against nil tag in error")
	}
}
