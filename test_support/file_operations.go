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
	_, err = f.WriteString("test contents")
	check(err)
	check(f.Close())
	return fp
}

func CreateDir(path string, dirName string) string {
	return CreateDirWithMode(path, dirName, os.FileMode(0777))
}

func CreateDirWithMode(path string, dirName string, mode os.FileMode) string {
	fp := filepath.Join(path, dirName)
	err := os.Mkdir(fp, mode)
	check(err)
	return fp
}

func CreatePrototype(baseDir string) string {
	pdir := CreateDir(baseDir, "test-prototype")
	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}
	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(pdir, dir), os.FileMode(0))
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
