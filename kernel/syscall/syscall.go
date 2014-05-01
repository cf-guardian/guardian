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
	/*
		Mounts the given source directory at the given mount point with the "bind" option.
	*/
	BindMountReadWrite(source string, mountPoint string) error

	/*
		Mounts the given source directory at the given mount point read-only with the "bind" option.
	*/
	BindMountReadOnly(source string, mountPoint string) error

	/*
		Unmounts the given mount point.
	*/
	Unmount(mountPoint string) error
}

type syscallWrapper struct {
}

func New() Syscall {
	return new(syscallWrapper)
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
