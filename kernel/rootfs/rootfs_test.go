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
	"errors"
	"github.com/cf-guardian/guardian/gerror"
	"github.com/cf-guardian/guardian/kernel/fileutils"
	"github.com/cf-guardian/guardian/kernel/fileutils/mock_fileutils"
	"github.com/cf-guardian/guardian/kernel/rootfs"
	"github.com/cf-guardian/guardian/kernel/syscall/mock_syscall"
	"github.com/cf-guardian/guardian/test_support"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNilSyscallFS(t *testing.T) {
	mockCtrl, mockFileUtils, _ := setupMocks(t)
	defer mockCtrl.Finish()

	rfs, gerr := rootfs.NewRootFS(nil, mockFileUtils, "")
	if rfs != nil || !gerr.EqualTag(rootfs.ErrNilSyscallFS) {
		t.Errorf("Incorrect return values (%s, %s)", rfs, gerr)
		return
	}
}

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

	tempDir := test_support.CreateTempDir()
	filePath := test_support.CreateFile(tempDir, "testFile")
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

	tempDir := test_support.CreateTempDir()
	dirPath := test_support.CreateDirWithMode(tempDir, "test-rootfs", os.FileMode(0400))
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

	tempDir := test_support.CreateTempDir()
	mockFileUtils.EXPECT().Filemode(tempDir).Return(os.ModeDir|os.FileMode(0700), nil)
	rfs, gerr := rootfs.NewRootFS(mockSyscallFS, mockFileUtils, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	prototypeDir := test_support.CreatePrototype(tempDir)

	mockSyscallFS.EXPECT().BindMountReadOnly(prototypeDir, test_support.NewStringPrefixMatcher(filepath.Join(tempDir, "mnt")))

	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}
	for _, dir := range dirs {
		srcMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, "tmp-rootfs-[^/]*", dir))
		mntMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, "mnt-[^/]*", dir))
		mockFileUtils.EXPECT().Exists(srcMatcher).Return(true).AnyTimes()
		mockFileUtils.EXPECT().Exists(mntMatcher).Return(true).AnyTimes()
		mockSyscallFS.EXPECT().BindMountReadWrite(srcMatcher, mntMatcher)
	}

	root, gerr := rfs.Generate(prototypeDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}

	rootPrefix := filepath.Join(tempDir, "mnt-")
	if !strings.HasPrefix(root, rootPrefix) {
		t.Errorf("root was %s, but expected it to have prefix %s", root, rootPrefix)
		return
	}
}

func TestGenerateBackoutAfterBindMountReadWriteError(t *testing.T) {
	for i := 0; i <= 6; i++ {
		testGenerateBackoutAfterBindMountReadWriteError(i, t)
	}
}

func testGenerateBackoutAfterBindMountReadWriteError(i int, t *testing.T) {
	mockCtrl, mockFileUtils, mockSyscallFS := setupMocks(t)
	defer mockCtrl.Finish()

	tempDir := test_support.CreateTempDir()
	mockFileUtils.EXPECT().Filemode(tempDir).Return(os.ModeDir|os.FileMode(0700), nil)
	rfs, gerr := rootfs.NewRootFS(mockSyscallFS, mockFileUtils, tempDir)
	if gerr != nil {
		t.Errorf("%s", gerr)
		return
	}
	prototypeDir := filepath.Join(tempDir, "test-prototype")

	mainMountPointMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, `mnt-[\d]*$`))
	mockSyscallFS.EXPECT().BindMountReadOnly(prototypeDir, mainMountPointMatcher)
	mockSyscallFS.EXPECT().Unmount(mainMountPointMatcher)

	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}
	for j := 0; j < i; j++ {
		dir := dirs[j]

		srcMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, "tmp-rootfs-[^/]*", dir))
		mntMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, "mnt-[^/]*", dir))
		mockFileUtils.EXPECT().Exists(srcMatcher).Return(true).AnyTimes()
		mockFileUtils.EXPECT().Exists(mntMatcher).Return(true).AnyTimes()
		mockSyscallFS.EXPECT().BindMountReadWrite(srcMatcher, mntMatcher)
		mockSyscallFS.EXPECT().Unmount(mntMatcher)
	}

	failingDir := dirs[i]
	srcMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, "tmp-rootfs-[^/]*", failingDir))
	mntMatcher := test_support.NewStringRegexMatcher(filepath.Join(tempDir, "mnt-[^/]*", failingDir))
	mockFileUtils.EXPECT().Exists(srcMatcher).Return(true).AnyTimes()
	mockFileUtils.EXPECT().Exists(mntMatcher).Return(true).AnyTimes()
	mockSyscallFS.EXPECT().BindMountReadWrite(srcMatcher, mntMatcher).Return(errors.New("test error"))

	root, gerr := rfs.Generate(prototypeDir)
	if gerr == nil {
		t.Errorf("Unexpected return values %s, %s", root, gerr)
		return
	}
}

func setupMocks(t *testing.T) (*gomock.Controller, *mock_fileutils.MockFileutils, *mock_syscall.MockSyscallFS) {
	mockCtrl := gomock.NewController(t)
	mockFileUtils := mock_fileutils.NewMockFileutils(mockCtrl)
	mockSyscallFS := mock_syscall.NewMockSyscallFS(mockCtrl)
	return mockCtrl, mockFileUtils, mockSyscallFS
}
