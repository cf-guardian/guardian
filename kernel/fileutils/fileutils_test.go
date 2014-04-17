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
	"path/filepath"
	"io/ioutil"
	"os"
	"testing"
)

func TestCopySingle(t *testing.T) {
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

func TestCopySingleMode(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	src := createFileWithMode(td, "src.file", os.FileMode(0642))
	target := filepath.Join(td, "target.file")
	err := fileutils.Copy(target, src)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	fi, err := os.Lstat(target)
	check(err)
	modeString := fi.Mode().String()
	expModeString := "-rw-r-----"
	if modeString != expModeString {
		t.Errorf("Copied file has incorrect file mode %q, expected %q", modeString, expModeString)
	}
}

func createFile(td string, fileName string) string {
	return createFileWithMode(td, fileName, os.FileMode(0666))
}

func createFileWithMode(td string, fileName string, mode os.FileMode) string {
	fp := filepath.Join(td, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE | os.O_EXCL | os.O_WRONLY, mode)
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

func checkFile(target string, expContents string, t *testing.T) {
	f, err := os.Open(target)
	check(err)
	buf := make([]byte, len(expContents))
	n, err := f.Read(buf)
	check(err)
	if actualContents := string(buf[:n]); actualContents != expContents {
		t.Errorf("Contents %q not expected value %q", actualContents, expContents)
	}
}

