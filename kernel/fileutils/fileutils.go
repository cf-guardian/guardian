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
	"os"
	"github.com/cf-guardian/guardian/gerror"
)

/*
	Copy copies a source file to a destination file. File contents are copied. File mode and permissions
	(as described in http://golang.org/pkg/os/#FileMode) are copied too.

	TODO: copy directories
	TODO: support symbolic links
	TODO: copy file owner
 */
func Copy(destPath string, srcPath string) error {
	return copyFile(destPath, srcPath)
}

func copyFile(destination string, source string) error {
	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		return gerror.FromError(err)
	}
	defer sourceFile.Close()

	fi, err := os.Lstat(source)
	if err != nil {
		return gerror.FromError(err)
	}

	destinationFile, err := os.OpenFile(destination, os.O_CREATE | os.O_EXCL | os.O_WRONLY, fi.Mode())
	if err != nil {
		return gerror.FromError(err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return gerror.FromError(err)
}
