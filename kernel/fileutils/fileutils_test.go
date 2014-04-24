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

package fileutils_test

import (
	"github.com/cf-guardian/guardian/kernel/fileutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	src := createFile(td, "src.file")
	target := filepath.Join(td, "target.file")
	err := fileutils.Copy(target, src)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	checkFile(target, "test contents", t)
}

func TestCopyNonExistent(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	badSrc := filepath.Join(td, "src.file")
	target := filepath.Join(td, "target.file")
	err := fileutils.Copy(target, badSrc)
	if err == nil {
		t.Errorf("Failed to return non-nil error")
		return
	}
}

func TestCopySameFile(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	src := createFile(td, "src.file")
	err := fileutils.Copy(src, src)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	checkFile(src, "test contents", t)
}

func TestCopyFileMode(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	src := createFileWithMode(td, "src.file", os.FileMode(0642))
	target := filepath.Join(td, "target.file")
	err := fileutils.Copy(target, src)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	modeString := fileMode(target).String()
	expModeString := "-rw-r-----"
	if modeString != expModeString {
		t.Errorf("Copied file has incorrect file mode %q, expected %q", modeString, expModeString)
	}
}

func TestCopyDirectoryToNew(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	srcDir := filepath.Join(td, "source")
	err := os.Mkdir(srcDir, os.FileMode(0777))
	check(err)

	createFile(srcDir, "file1")
	createFile(srcDir, "file2")

	targetDir := filepath.Join(td, "target")
	err = fileutils.Copy(targetDir, srcDir)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	checkDirectory(targetDir, t)
	checkFile(filepath.Join(targetDir, "file1"), "test contents", t)
	checkFile(filepath.Join(targetDir, "file2"), "test contents", t)
}

func TestCopyDirectoryNestedToNew(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	srcDir := filepath.Join(td, "source")
	err := os.Mkdir(srcDir, os.FileMode(0777))
	check(err)

	subDir := filepath.Join(srcDir, "subdir")
	err = os.Mkdir(subDir, os.FileMode(0777))
	check(err)

	createFile(subDir, "file1")
	createFile(subDir, "file2")

	targetDir := filepath.Join(td, "target")
	err = fileutils.Copy(targetDir, srcDir)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	checkDirectory(targetDir, t)
	checkFile(filepath.Join(targetDir, "subdir", "file1"), "test contents", t)
	checkFile(filepath.Join(targetDir, "subdir", "file2"), "test contents", t)
}

func TestCopyDirectoryToExisting(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	srcDir := filepath.Join(td, "source")
	err := os.Mkdir(srcDir, os.FileMode(0777))
	check(err)

	createFile(srcDir, "file1")
	createFile(srcDir, "file2")

	targetDir := filepath.Join(td, "target")
	err = os.Mkdir(targetDir, os.FileMode(0777))
	check(err)
	err = fileutils.Copy(targetDir, srcDir)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}

	resultantDir := filepath.Join(targetDir, "source")
	checkDirectory(resultantDir, t)
	checkFile(filepath.Join(resultantDir, "file1"), "test contents", t)
	checkFile(filepath.Join(resultantDir, "file2"), "test contents", t)
}

func TestCopyDirMode(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	src := createDirWithMode(td, "src.dir", os.FileMode(0642))
	target := filepath.Join(td, "target.dir")
	err := fileutils.Copy(target, src)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	modeString := fileMode(target).String()
	expModeString := "drw-r-----"
	if modeString != expModeString {
		t.Errorf("Copied directory has incorrect file mode %q, expected %q", modeString, expModeString)
	}
}

func createDirWithMode(td string, dirName string, mode os.FileMode) string {
	fp := filepath.Join(td, dirName)
	err := os.Mkdir(fp, mode)
	check(err)
	return fp
}

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

func createTmpDir() string {
	tPath, err := ioutil.TempDir("/tmp", "fileutils_test-")
	check(err)
	return tPath
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkDirectory(path string, t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("Not a directory: %q (%v)", path, e)
		}
	}()

	if !fileMode(path).IsDir() {
		t.Errorf("Not a directory: %q", path)
	}
}

func checkFile(target string, expContents string, t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("Problem with file: %q (%v)", target, e)
		}
	}()

	f, err := os.Open(target)
	check(err)
	buf := make([]byte, len(expContents))
	n, err := f.Read(buf)
	check(err)
	if actualContents := string(buf[:n]); actualContents != expContents {
		t.Errorf("Contents %q not expected value %q", actualContents, expContents)
	}
}

func fileMode(path string) os.FileMode {
	fi, err := os.Lstat(path)
	check(err)
	return fi.Mode()
}
