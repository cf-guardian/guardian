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

package test_support

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func CreateTempDir() string {
	tempDir, err := ioutil.TempDir("/tmp", "guardian-test-")
	check(err)
	return tempDir
}

func CreateFile(path string, fileName string) string {
	return CreateFileWithMode(path, fileName, os.FileMode(0666))
}

func CreateFileWithMode(path string, fileName string, mode os.FileMode) string {
	fp := filepath.Join(path, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	check(err)
	defer f.Close()
	_, err = f.WriteString("test contents")
	check(err)
	return fp
}

// Create a file and return any error.
func TestCreateFile(t *testing.T, td string, fileName string) (_ string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return CreateFile(td, fileName), nil
}

// Create a file and return any error.
func TestCreateFileWithMode(t *testing.T, td string, fileName string, mode os.FileMode) (_ string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return CreateFileWithMode(td, fileName, mode), nil
}


func CreateDir(path string, dirName string) string {
	return CreateDirWithMode(path, dirName, os.FileMode(0755))
}

func CreateDirWithMode(path string, dirName string, mode os.FileMode) string {
	fp := filepath.Join(path, dirName)
	err := os.Mkdir(fp, mode)
	check(err)
	return fp
}

func RootFSDirs() []string {
	return []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}
}

func CreatePrototype(baseDir string) string {
	pdir := CreateDir(baseDir, "test-prototype")
	for _, dir := range RootFSDirs() {
		os.MkdirAll(filepath.Join(pdir, dir), os.FileMode(0755))
	}
	return pdir
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func SameFile(p1 string, p2 string) bool {
	fi1, err := os.Stat(p1)
	check(err)
	fi2, err := os.Stat(p2)
	check(err)
	return os.SameFile(fi1, fi2)
}

func FileMode(path string) os.FileMode {
	fi, err := os.Lstat(path)
	check(err)
	return fi.Mode()
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func CleanupDirs(t *testing.T, paths... string) {
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("Could not delete %s", path)
		}
	}
}
