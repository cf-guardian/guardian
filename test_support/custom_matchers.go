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

package test_support

import (
	"fmt"
	"github.com/golang/glog"
	"regexp"
	"strings"
)

func NewStringPrefixMatcher(prefix string) *stringPrefixMatcher {
	return &stringPrefixMatcher{prefix}
}

type stringPrefixMatcher struct {
	prefix string
}

func (m *stringPrefixMatcher) Matches(x interface{}) bool {
	if x, ok := x.(string); ok {
		matched := strings.HasPrefix(x, m.prefix)
		if glog.V(2) {
			glog.Infof("Result of HasPrefix(%q, %q) is %v", x, m.prefix, matched)
		}
		return matched
	} else {
		glog.Errorf("Not a string: %v", x)
		return false
	}
}

func (m *stringPrefixMatcher) String() string {
	return fmt.Sprintf("is a string with prefix %s", m.prefix)
}

func NewStringRegexMatcher(regex string) *stringRegexMatcher {
	return &stringRegexMatcher{regex}
}

type stringRegexMatcher struct {
	regex string
}

func (m *stringRegexMatcher) Matches(x interface{}) bool {
	if x, ok := x.(string); ok {
		if matched, err := regexp.MatchString(m.regex, x); err == nil {
			if glog.V(2) {
				glog.Infof("Result of MatchString(%q, %q) is %v", m.regex, x, matched)
			}
			return matched
		} else {
			glog.Errorf("Invalid regular expression %q: %s", m.regex, err)
			return false
		}
	} else {
		glog.Errorf("Not a string: %v", x)
		return false
	}
}

func (m *stringRegexMatcher) String() string {
	return fmt.Sprintf("is a string which matches regular expression %s", m.regex)
}
