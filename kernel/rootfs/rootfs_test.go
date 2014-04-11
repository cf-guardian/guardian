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

package rootfs_test

import (
	"ioutil"
	"testing"
	"github.com/cf-guardian/guardian/kernel/syscall"
	"github.com/cf-guardian/guardian/kernel/rootfs""
)

type stubSyscall struct {
	callCount int
}

func (ss *stubSyscall) Mount(source string, target string, fstype string, flags uintptr, data string) error {
	return nil
}

func TestGenerate(t *testing.T) {
	prototype, err := ioutil.TempDir("/tmp/guardian", "test-rootfs")
	if err != nil {
		t.Errorf("%q", err)
	}

	_, err := rootfs.Generate(prototype, &stubSyscall{})
	if err != nil {
		t.Errorf("%q", err)
	}	
}
