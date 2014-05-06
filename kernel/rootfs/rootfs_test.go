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
	"code.google.com/p/gomock/gomock"
	"fmt"
	"github.com/cf-guardian/guardian/gerror"
	"github.com/cf-guardian/guardian/kernel/fileutils"
	"github.com/cf-guardian/guardian/kernel/fileutils/mock_fileutils"
	"github.com/cf-guardian/guardian/kernel/rootfs"
	"github.com/cf-guardian/guardian/kernel/syscall/mock_syscall"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNonExistentReadWriteBaseDir(t *testing.T) {
	mockCtrl, mockFileUtils, mockSyscallFS := setupMocks(t)
	defer mockCtrl.Finish()

	mockFileUtils.EXPECT().Filemode("/nosuch").Return(os.FileMode(0), gerror.New(fileutils.ErrFileNotFound, "test error"))

	rfs, gerr := rootfs.NewRootFS(mockSyscallFS, mockFileUtils, "/nosuch")
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirMissing) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestNonDirReadWriteBaseDir(t *testing.T) {
	mockCtrl, mockFileUtils, mockSyscallFS := setupMocks(t)
	defer mockCtrl.Finish()

	tempDir := createTempDir()
	filePath := createFile(tempDir, "testFile")
	mockFileUtils.EXPECT().Filemode(filePath).Return(os.FileMode(0700), nil)

	rfs, gerr := rootfs.NewRootFS(mockSyscallFS, mockFileUtils, filePath)
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirIsFile) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestReadOnlyReadWriteBaseDir(t *testing.T) {
	mockCtrl, mockFileUtils, mockSyscallFS := setupMocks(t)
	defer mockCtrl.Finish()

	tempDir := createTempDir()
	dirPath := createDirWithMode(tempDir, "test-rootfs", os.FileMode(0400))
	mockFileUtils.EXPECT().Filemode(dirPath).Return(os.ModeDir|os.FileMode(0100), nil)

	rfs, gerr := rootfs.NewRootFS(mockSyscallFS, mockFileUtils, dirPath)
	if rfs != nil || !gerr.EqualTag(rootfs.ErrRwBaseDirNotRw) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

func TestGenerate(t *testing.T) {
	mockCtrl, mockFileUtils, mockSyscallFS := setupMocks(t)
	defer mockCtrl.Finish()

	tempDir := createTempDir()
	mockFileUtils.EXPECT().Filemode(tempDir).Return(os.ModeDir|os.FileMode(0700), nil)
	rfs, gerr := rootfs.NewRootFS(mockSyscallFS, mockFileUtils, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	prototypeDir := filepath.Join(tempDir, "test-prototype")

	mockSyscallFS.EXPECT().BindMountReadOnly(prototypeDir, &stringPrefixMatcher{filepath.Join(tempDir, "mnt")})

	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}
	for _, dir := range dirs {
		mockSyscallFS.EXPECT().BindMountReadWrite(&stringRegexMatcher{filepath.Join(tempDir, "tmp-rootfs-.*", dir)}, &stringRegexMatcher{filepath.Join(tempDir, "mnt-.*", dir)})
	}

	root, gerr := rfs.Generate(prototypeDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}

	_ = root
}

type stringPrefixMatcher struct {
	prefix string
}

func (m *stringPrefixMatcher) Matches(x interface{}) bool {
	if x, ok := x.(string); ok {
		return strings.HasPrefix(x, m.prefix)
	} else {
		return false
	}
}

func (m *stringPrefixMatcher) String() string {
	return fmt.Sprintf("is a string with prefix %s", m.prefix)
}

type stringRegexMatcher struct {
	regex string
}

func (m *stringRegexMatcher) Matches(x interface{}) bool {
	if x, ok := x.(string); ok {
		if matched, err := regexp.MatchString(m.regex, x); err == nil {
			return matched
		} else {
			return false
		}
	} else {
		return false
	}
}

func (m *stringRegexMatcher) String() string {
	return fmt.Sprintf("is a string which matches regular expression %s", m.regex)
}

func createTempDir() string {
	tempDir, err := ioutil.TempDir("/tmp", "guardian-test-")
	check(err)
	return tempDir
}

func setupMocks(t *testing.T) (*gomock.Controller, *mock_fileutils.MockFileutils, *mock_syscall.MockSyscall_FS) {
	mockCtrl := gomock.NewController(t)
	mockFileUtils := mock_fileutils.NewMockFileutils(mockCtrl)
	mockSyscallFS := mock_syscall.NewMockSyscall_FS(mockCtrl)
	return mockCtrl, mockFileUtils, mockSyscallFS
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
