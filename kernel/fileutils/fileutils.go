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

/*
Package fileutils provides some file manipulation utilities.
*/
package fileutils

import (
	"io"
	"log"
	"os"
	"github.com/cf-guardian/guardian/gerror"
)

func Copy(destPath string, srcPath string) error {
	assertExists(srcPath)
	err := copyFile(destPath, srcPath)
	assertExists(destPath)
	return err
}

func assertExists(f string) {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		panic(err)
	}
}

func copyFile(destination string, source string) error {
	log.Printf("copyFile(%s, %s)\n", destination, source)

	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		log.Printf("copyFile(%s, %s) gave error %s\n", destination, source, err)
		return gerror.FromError(err)
	}
	defer sourceFile.Close()

//	fi, err := os.Lstat(source)
//	if err != nil {
//		return gerror.FromError(err)
//	}

	destinationFile, err := os.Create(destination) // os.OpenFile(destination, os.O_CREATE + os.O_EXCL, fi.Mode())
	if err != nil {
		return gerror.FromError(err)
	}
	defer destinationFile.Close()

	n, err := io.Copy(destinationFile, sourceFile)
	gerr := gerror.FromError(err)
	log.Printf("copyFile(%s, %s) copied %d bytes and returning %v -> %v\n", destination, source, n, err, gerr)

	return gerr
}
