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

package syscall_test

import (
	"github.com/cf-guardian/guardian/kernel/syscall"
	"github.com/cf-guardian/guardian/kernel/syscall/syscall_linux"
	"github.com/cf-guardian/guardian/test_support"
	"testing"
	"path/filepath"
)

func TestBindMountReadWrite(t *testing.T) {
	sc := setup(t)
	dir := test_support.CreateTempDir()
	mountPoint := test_support.CreateTempDir()
	err := sc.BindMountReadWrite(dir, mountPoint)
	if err != nil {
		t.Errorf("BindMountReadWrite failed: %s", err)
	}

	err = sc.Unmount(mountPoint)
	if err != nil {
		t.Errorf("Unmount failed: %s", err)
	}
}

func TestBindMountReadOnly(t *testing.T) {
	sc := setup(t)
	dir := test_support.CreateTempDir()
	mountPoint := test_support.CreateTempDir()
	err := sc.BindMountReadOnly(dir, mountPoint)
	if err != nil {
		t.Errorf("BindMountReadOnly failed: %s", err)
	}

	err = sc.Unmount(mountPoint)
	if err != nil {
		t.Errorf("Unmount failed: %s", err)
	}
}

func TestBindMountOverlay(t *testing.T) {
	sc := setup(t)
	dir := test_support.CreateTempDir()
	subdir := test_support.CreateDir(dir, "subdir")
	test_support.CreateFile(subdir, "test.file")
	mountPoint := test_support.CreateTempDir()

	err := sc.BindMountReadOnly(dir, mountPoint)
	if err != nil {
		t.Errorf("BindMountReadOnly failed: %s", err)
	}

	overlay := test_support.CreateTempDir()
	overlayMountPoint := filepath.Join(mountPoint, "subdir")
	err = sc.BindMountReadWrite(overlay, overlayMountPoint)
	if err != nil {
		t.Errorf("BindMountReadWrite failed: %s", err)
	}

	err = sc.Unmount(overlayMountPoint)
	if err != nil {
		t.Errorf("Unmount %s failed: %s", overlayMountPoint, err)
	}

	err = sc.Unmount(mountPoint)
	if err != nil {
		t.Errorf("Unmount %s failed: %s", mountPoint, err)
	}
}

func setup(t *testing.T) syscall.SyscallFS {
	sc, err := syscall_linux.NewFS()
	if err != nil {
		t.Error("SyscallFS requires root privileges - run the test as root")
		panic("Test aborted, must be run as root")
	}
	return sc
}
