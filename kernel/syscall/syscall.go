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
	BindMount(source string, mountPoint string, flags... uintptr) error
	Unmount(mountPoint string) error
}

type syscallWrapper struct {
}

func New() Syscall {
	return new(syscallWrapper)
}

const NO_FLAGS uintptr = 0
const MS_RDONLY uintptr = trueSyscall.MS_RDONLY

func (_ *syscallWrapper) BindMount(source string, mountPoint string, flags... uintptr) error {
	var fl uintptr = uintptr(trueSyscall.MS_BIND)
	for _, f := range flags {
		fl = fl + f
	}
	return trueSyscall.Mount(source, mountPoint, "", fl, "")
}

func (_ *syscallWrapper) Unmount(mountPoint string) error {
	return trueSyscall.Unmount(mountPoint, 0)
}
