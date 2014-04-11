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
Package syscall wraps the standard syscall package to support testing.
*/
package syscall

import (
	trueSyscall "syscall"
)

type Syscall interface {
	Mount(source string, target string, fstype string, flags uintptr, data string) error
}

type syscallWrapper struct {
}

func (scw *syscallWrapper) Mount(source string, target string, fstype string, flags uintptr, data string) error {
	return trueSyscall.Mount(source, target, fstype, flags, data)
}
