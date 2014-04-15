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
	"log"
	"os"
	"reflect"
	"testing"
)

func TestCopySingle(t *testing.T) {
	td := createTmpDir()
	defer os.RemoveAll(td)

	src := createFile(td, "src.file")
	target := filepath.Join(td, "target.file")
	err := fileutils.Copy(target, src)
	if err != nil {
		log.Printf("Type of err = %s; value = %v", reflect.TypeOf(err), err)
		t.Errorf("Failed: %s %v", err, nil)
		return
	}
	checkFile(target, "test contents", t)
}

func createFile(td string, fileName string) string {
	fp := filepath.Join(td, fileName)
	f, err := os.Create(fp)
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

