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
Package syscall-linux wraps the standard syscall package for Linux.
*/
package syscall_linux

import (
	"github.com/cf-guardian/guardian/gerror"
	syscall "github.com/cf-guardian/guardian/kernel/syscall"
	"os"
	trueSyscall "syscall"
)

// ImplErrorId is used for error ids relating to the implementation of this package.
type ImplErrorId int

const (
	ErrNotRoot ImplErrorId = iota // root is required to create a SyscallFS
)

type syscallWrapper struct {
}

/*
	Constructs a new SyscallFS instance and returns it providing the effective user id
	is root. Otherwise return an error.
*/
func NewFS() (syscall.SyscallFS, error) {
	euid := os.Geteuid()
	if euid != 0 {
		return nil, gerror.Newf(ErrNotRoot, "Effective user id %d is not root", euid)
	}
	return &syscallWrapper{}, nil
}

func (_ *syscallWrapper) BindMountReadWrite(source string, mountPoint string) error {
	return trueSyscall.Mount(source, mountPoint, "", trueSyscall.MS_BIND, "")
}

func (_ *syscallWrapper) BindMountReadOnly(source string, mountPoint string) error {

	return trueSyscall.Mount(source, mountPoint, "", trueSyscall.MS_BIND|trueSyscall.MS_RDONLY, "")
}

func (_ *syscallWrapper) Unmount(mountPoint string) error {
	return trueSyscall.Unmount(mountPoint, 0)
}
