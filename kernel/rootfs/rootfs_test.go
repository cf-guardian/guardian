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
	"fmt"
	"github.com/cf-guardian/guardian/kernel/rootfs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type stubSyscall struct {
	callCount int
}

func (ss *stubSyscall) BindMountReadOnly(source string, mountPoint string) error {
	return nil
}

func (ss *stubSyscall) BindMountReadWrite(source string, mountPoint string) error {
	return nil
}

func (ss *stubSyscall) Unmount(mountPoint string) error {
	return nil
}

func TestNonExistentReadWriteBaseDir(t *testing.T) {
	rfs, gerr := rootfs.NewRootFS(&stubSyscall{}, "/nosuch")
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirMissing) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestNonDirReadWriteBaseDir(t *testing.T) {
	tempDir := createTempDir()
	filePath := createFile(tempDir, "testFile")

	rfs, gerr := rootfs.NewRootFS(&stubSyscall{}, filePath)
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirIsFile) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestReadOnlyReadWriteBaseDir(t *testing.T) {
	tempDir := createTempDir()
	dirPath := createDirWithMode(tempDir, "test-rootfs", os.FileMode(0400))

	rfs, gerr := rootfs.NewRootFS(&stubSyscall{}, dirPath)
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirNotRw) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestGenerate(t *testing.T) {
	tempDir := createTempDir()
	rfs, gerr := rootfs.NewRootFS(&stubSyscall{}, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	os.MkdirAll("/tmp/guardian-test", 0700)
	prototype, err := ioutil.TempDir("/tmp/guardian-test", "test-rootfs")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	os.MkdirAll(prototype, 0700)

	root, err := rfs.Generate(prototype)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	fmt.Println(root)
}

func createTempDir() string {
	tempDir, err := ioutil.TempDir("/tmp", "guardian-test")
	check(err)
	return tempDir
}

// TODO: Remove duplication with fileutils_test.
func createFile(td string, fileName string) string {
	return createFileWithMode(td, fileName, os.FileMode(0666))
}

func createFileWithMode(td string, fileName string, mode os.FileMode) string {
	fp := filepath.Join(td, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	check(err)
	_, err = f.WriteString("test contents")
	check(err)
	check(f.Close())
	return fp
}

func createDir(td string, dirName string) string {
	return createDirWithMode(td, dirName, os.FileMode(0777))
}

func createDirWithMode(td string, dirName string, mode os.FileMode) string {
	fp := filepath.Join(td, dirName)
	err := os.Mkdir(fp, mode)
	check(err)
	return fp
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
