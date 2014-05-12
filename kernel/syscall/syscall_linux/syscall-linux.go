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
	"github.com/golang/glog"
	"os"
	trueSyscall "syscall"
	"io/ioutil"
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

func (sc *syscallWrapper) BindMountReadOnly(source string, mountPoint string) error {
	// On Linux, a read-only bind mount may turn out read-write and must be remounted to make it read-only.
	err := doBindMountReadOnly(source, mountPoint)
	if err != nil {
		return err
	}
	readOnly := checkReadOnly(mountPoint)
	if !readOnly {
		if glog.V(2) {
			glog.Infof("Remounting bind mount %s read-only", mountPoint)
		}
		err = doBindRemountReadOnly(source, mountPoint)
		if err != nil {
			if unmountErr := sc.Unmount(mountPoint); unmountErr != nil {
				glog.Warningf("Failed to undo bind mount of %s while recovering from %s", mountPoint, err)
			}
			return err
		}
		if !checkReadOnly(mountPoint) {
			glog.Warningf("Failed to remount bind mount of %s read-only", mountPoint)
			if unmountErr := sc.Unmount(mountPoint); unmountErr != nil {
				glog.Warningf("Failed to undo bind mount of %s while recovering from failure to remount read-only", mountPoint)
			}
			return gerror.Newf("Failed to remount bind mount of %s read-only", mountPoint)
		} else {
			if glog.V(2) {
				glog.Infof("Successfully remounted bind mount %s read-only", mountPoint)
			}
		}
	}
	return nil
}

func doBindMountReadOnly(source string, mountPoint string) error {
	return trueSyscall.Mount(source, mountPoint, "", trueSyscall.MS_BIND|trueSyscall.MS_RDONLY, "")
}

func doBindRemountReadOnly(source string, mountPoint string) error {
	return trueSyscall.Mount(source, mountPoint, "", trueSyscall.MS_BIND|trueSyscall.MS_REMOUNT|trueSyscall.MS_RDONLY, "")
}

func checkReadOnly(mountPoint string) bool {
	path, err := ioutil.TempDir(mountPoint, "cf-guardian-ro-check-")
	if err != nil {
		return true
	} else {
		err = os.Remove(path)
		if err != nil {
			glog.Warningf("Failed to delete file %s used to check read-only bind mount", path)
		}
		return false
	}
}

func (_ *syscallWrapper) Unmount(mountPoint string) error {
	return trueSyscall.Unmount(mountPoint, 0)
}
