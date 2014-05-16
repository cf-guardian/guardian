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
	"github.com/cf-guardian/guardian/kernel/fileutils"
	"github.com/cf-guardian/guardian/kernel/rootfs"
	"github.com/cf-guardian/guardian/kernel/syscall"
	"github.com/cf-guardian/guardian/kernel/syscall/syscall_linux"
	"github.com/cf-guardian/guardian/test_support"
	"os"
	"path/filepath"
	"testing"
)

func TestNonExistentReadWriteBaseDir(t *testing.T) {
	syscallFS, futils := setup(t)
	rfs, gerr := rootfs.NewRootFS(syscallFS, futils, "/nosuch")
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirMissing) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestNonDirReadWriteBaseDir(t *testing.T) {
	syscallFS, futils := setup(t)

	tempDir := test_support.CreateTempDir()
	defer test_support.CleanupDirs(t, tempDir)

	filePath := test_support.CreateFile(tempDir, "testFile")

	rfs, gerr := rootfs.NewRootFS(syscallFS, futils, filePath)
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirIsFile) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestReadOnlyReadWriteBaseDir(t *testing.T) {
	syscallFS, futils := setup(t)

	tempDir := test_support.CreateTempDir()
	defer test_support.CleanupDirs(t, tempDir)

	dirPath := test_support.CreateDirWithMode(tempDir, "test-rootfs", os.FileMode(0400))

	rfs, gerr := rootfs.NewRootFS(syscallFS, futils, dirPath)
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirNotRw) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestGenerateMissingRootSubdir(t *testing.T) {
	syscallFS, futils := setup(t)

	tempDir := test_support.CreateTempDir()
	defer test_support.CleanupDirs(t, tempDir)

	rfs, gerr := rootfs.NewRootFS(syscallFS, futils, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	prototypeDir := test_support.CreatePrototype(tempDir)
	os.Remove(filepath.Join(prototypeDir, `home`))

	_, gerr = rfs.Generate(prototypeDir)
	if gerr == nil || !gerr.EqualTag(rootfs.ErrRootSubdirMissing){
		t.Errorf("Incorrect error %s", gerr)
		return
	}
}

func TestGenerate(t *testing.T) {
	syscallFS, futils := setup(t)

	tempDir := test_support.CreateTempDir()
	defer test_support.CleanupDirs(t, tempDir)

	rfs, gerr := rootfs.NewRootFS(syscallFS, futils, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	prototypeDir := test_support.CreatePrototype(tempDir)

	// home directory should be copied, so give it some content.
	test_support.CreateFile(filepath.Join(prototypeDir, "home"), "test.home")

	// tmp directory should not be copied, so give it some content.
	test_support.CreateFile(filepath.Join(prototypeDir, "tmp"), "test.tmp")

	root, gerr := rfs.Generate(prototypeDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}

	checkRootFS(root, prototypeDir, t)

	// Check home directory copied and writeable.
	homePath := filepath.Join(root, "home", "test.home")
	if !test_support.FileExists(homePath) {
		t.Errorf("home/test.home does not exist in generated root filesystem")
	}
	err := os.Remove(homePath)
	if err != nil {
		t.Errorf("Failed to delete file from home directory of root %s: %s", root, err)
	}

	// Check tmp directory was not copied.
	if test_support.FileExists(filepath.Join(root, "tmp", "test.tmp")) {
		t.Errorf("tmp/test.tmp should not exist in generated root filesystem")
	}

	gerr = rfs.Remove(root)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}

	if test_support.FileExists(root) {
		t.Errorf("root %s was not unmounted", root)
		return
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Errorf("Error removing test directory %s", err)
		return
	}
}

func checkRootFS(root string, prototypeDir string, t *testing.T) {
	_, err := test_support.TestCreateFile(t, root, "test.root")
	if err == nil {
		t.Errorf("Created file in read-only section of root %s", root)
	}
	path, err := test_support.TestCreateFile(t, filepath.Join(root, "tmp"), "test.write")
	if err != nil {
		t.Errorf("Failed to create file in tmp directory of root %s: %s", root, err)
	}
	err = os.Remove(path)
	if err != nil {
		t.Errorf("Failed to delete file from tmp directory of root %s: %s", root, err)
	}
}

func setup(t *testing.T) (syscall.SyscallFS, fileutils.Fileutils) {
	sc, err := syscall_linux.NewFS()
	if err != nil {
		t.Error("SyscallFS requires root privileges - run the test as root")
		panic("Test aborted, must be run as root")
	}
	return sc, fileutils.New()
}
