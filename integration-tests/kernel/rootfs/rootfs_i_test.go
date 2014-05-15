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
	rfs, gerr := rootfs.NewRootFS(syscallFS, futils, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	prototypeDir := test_support.CreatePrototype(tempDir)

	root, gerr := rfs.Generate(prototypeDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}

	checkRootFS(root, prototypeDir, t)

	gerr = rfs.Remove(root)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}

	err := os.RemoveAll(tempDir)
	if err != nil {
		t.Errorf("Error removing test directory %s", err)
		return
	}
}

func checkRootFS(root string, prototypeDir string, t *testing.T) {
	_, err := createFile(root, "test")
	if err == nil {
		t.Errorf("Created file in read-only section of root %s", root)
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

func createFile(td string, fileName string) (string, error) {
	return createFileWithMode(td, fileName, os.FileMode(0666))
}

func createFileWithMode(td string, fileName string, mode os.FileMode) (string, error) {
	fp := filepath.Join(td, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return "", err
	}
	_, err = f.WriteString("test contents")
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	return fp, nil
}
