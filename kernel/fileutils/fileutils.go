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
	"github.com/cf-guardian/guardian/gerror"
	"github.com/golang/glog"
	"io"
	"os"
	"path/filepath"
)

/*
	Copy copies a source file to a destination file. File contents are copied. File mode and permissions
	(as described in http://golang.org/pkg/os/#FileMode) are copied too.

	TODO: support symbolic links
	TODO: copy file owner
*/
func Copy(destPath string, srcPath string) error {
	glog.Infof("Copy(%s, %s)", destPath, srcPath)
	srcMode, gerr := fileMode(srcPath)
	if gerr != nil {
		return gerr
	}

	if srcMode.IsDir() {
		return copyDir(destPath, srcPath)
	} else {
		return copyFile(destPath, srcPath)
	}
}

func copyDir(destination string, source string) error {
	glog.Infof("copyDir(%s, %s)", destination, source)
	finalDestination, err := finalDestinationDir(destination, source)
	if err != nil {
		return gerror.FromError(err)
	}
	src, err := os.Open(source)
	if err != nil {
		return gerror.FromError(err)
	}

	names, err := src.Readdirnames(-1)
	if err != nil {
		return gerror.FromError(err)
	}

	for _, name := range names {
		glog.Infof("copying %s from %s to %s", name, source, finalDestination)
		err = Copy(filepath.Join(finalDestination, name), filepath.Join(source, name))
		if err != nil {
			return gerror.FromError(err)
		}
	}

	return nil
}

/*
	Determine the final destination directory and return an opened file referring to it.
*/
func finalDestinationDir(destination string, source string) (finalDestination string, err error) {
	defer func() {
		glog.Infof("openFinalDestinationDir(%s, %s) returning (%v, %v)", destination, source, finalDestination, err)
	}()
	sourceMode, err := fileMode(source)
	if err != nil {
		return finalDestination, gerror.FromError(err)
	}
	_, err = os.Stat(destination)
	if err != nil {
		if !os.IsNotExist(err) {
			return finalDestination, gerror.FromError(err)
		}
		finalDestination = destination
	} else {
		finalDestination = filepath.Join(destination, filepath.Base(source))
	}
	err = os.MkdirAll(finalDestination, sourceMode)
	if err != nil {
		return finalDestination, gerror.FromError(err)
	}
	return finalDestination, err
}

func copyFile(destination string, source string) error {
	glog.Infof("copyFile(%s, %s)", destination, source)
	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		return gerror.FromError(err)
	}
	defer sourceFile.Close()

	mode, gerr := fileMode(source)
	if gerr != nil {
		return gerr
	}

	destinationFile, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return gerror.FromError(err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return gerror.FromError(err)
}

func fileMode(path string) (os.FileMode, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return os.FileMode(0), gerror.FromError(err)
	}
	return fi.Mode(), nil
}
