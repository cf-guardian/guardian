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
	"strings"
)

type ErrorId int

const (
	ErrFileNotFound ErrorId = iota
	ErrOpeningSourceDir
	ErrCannotListSourceDir
	ErrUnexpected
	ErrCreatingTargetDir
	ErrOpeningSourceFile
	ErrOpeningTargetFile
	ErrCopyingFile
	ErrReadingSourceSymlink
	ErrWritingTargetSymlink
	ErrExternalSymlink
)

type Fileutils interface {

	/*
		Copy copies a source file to a destination file. File contents are copied. File mode and permissions
		(as described in http://golang.org/pkg/os/#FileMode) are copied.

		Directories are copied, along with their contents.

		Copying a file or directory to itself succeeds but does not modify the filesystem.

		Symbolic links are not followed and are copied provided they refer to a file or directory being copied
		(otherwise a non-nil error is returned). The only exception is copying a symbolic link to itself, which
		always succeeds.
	*/
	Copy(destPath string, srcPath string) gerror.Gerror

	/*
		Tests the existence of a file or directory at a given path. Returns true if and only if the file or
		directory exists.
	 */
	Exists(path string) bool

	/*
		Filemode returns the os.FileMode of the file with the given path. If the file does not exist, returns
		an error with tag ErrFileNotFound.
	*/
	Filemode(path string) (os.FileMode, gerror.Gerror)
}

type futils struct {
}

func New() Fileutils {
	return &futils{}
}

func (f *futils) Copy(destPath string, srcPath string) gerror.Gerror {
	if glog.V(1) {
		glog.Infof("Copy(%q, %q)", destPath, srcPath)
	}
	return f.doCopy(destPath, srcPath, srcPath)
}

func (f *futils) doCopy(destPath string, srcPath string, topSrcPath string) gerror.Gerror {
	if f.sameFile(srcPath, destPath) {
		return nil
	}
	srcMode, gerr := f.Filemode(srcPath)
	if gerr != nil {
		return gerr
	}

	if srcMode&os.ModeSymlink == os.ModeSymlink {
		return f.copySymlink(destPath, srcPath, topSrcPath)
	} else if srcMode.IsDir() {
		return f.copyDir(destPath, srcPath, topSrcPath)
	} else {
		return f.copyFile(destPath, srcPath)
	}
}

func (f *futils) copyDir(destination string, source string, topSource string) gerror.Gerror {
	if glog.V(2) {
		glog.Infof("copyDir(%q, %q)", destination, source)
	}
	finalDestination, gerr := f.finalDestinationDir(destination, source)
	if gerr != nil {
		return gerr
	}

	names, gerr := getNames(source)
	if gerr != nil {
		return gerr
	}

	for _, name := range names {
		if glog.V(2) {
			glog.Infof("copying %q from %q to %q", name, source, finalDestination)
		}
		gerr = f.doCopy(filepath.Join(finalDestination, name), filepath.Join(source, name), topSource)
		if gerr != nil {
			return gerr
		}
	}

	return nil
}

func getNames(dirPath string) (names [] string, gerr gerror.Gerror) {
	src, err := os.Open(dirPath)
	if err != nil {
		return names, gerror.NewFromError(ErrOpeningSourceDir, err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			glog.Warningf("Cannot close %v", src)
		}
	}()

	names, err = src.Readdirnames(-1)
	if err != nil {
		return names, gerror.NewFromError(ErrCannotListSourceDir, err)
	}

	return names, nil
}

/*
	Determine the final destination directory and return an opened file referring to it.
*/
func (f *futils) finalDestinationDir(destination string, source string) (finalDestination string, gerr gerror.Gerror) {
	if glog.V(2) {
		glog.Infof("openFinalDestinationDir(%q, %q)", destination, source)
		defer func() {
			glog.Infof("openFinalDestinationDir(%q, %q) returning (%v, %v)", destination, source, finalDestination, gerr)
		}()
	}
	sourceMode, gerr := f.Filemode(source)
	if gerr != nil {
		return finalDestination, gerr
	}
	if _, err := os.Stat(destination); err != nil {
		if !os.IsNotExist(err) {
			return finalDestination, gerror.NewFromError(ErrUnexpected, err)
		}
		finalDestination = destination
	} else {
		finalDestination = filepath.Join(destination, filepath.Base(source))
	}
	if err := os.MkdirAll(finalDestination, sourceMode); err != nil {
		finalDestination = ""
		return finalDestination, gerror.NewFromError(ErrCreatingTargetDir, err)
	}
	return finalDestination, nil
}

func (f *futils) copyFile(destination string, source string) gerror.Gerror {
	if glog.V(2) {
		glog.Infof("copyFile(%q, %q)", destination, source)
	}
	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		return gerror.NewFromError(ErrOpeningSourceFile, err)
	}
	defer sourceFile.Close()

	mode, gerr := f.Filemode(source)
	if gerr != nil {
		return gerr
	}

	destinationFile, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return gerror.NewFromError(ErrOpeningTargetFile, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return gerror.NewFromError(ErrCopyingFile, err)
}

func (f *futils) copySymlink(destLinkPath string, srcLinkPath string, topSrcPath string) gerror.Gerror {
	if glog.V(2) {
		glog.Infof("copySymLink(%q, %q, %q)", destLinkPath, srcLinkPath, topSrcPath)
	}
	linkTarget, err := os.Readlink(srcLinkPath)
	if err != nil {
		return gerror.NewFromError(ErrReadingSourceSymlink, err)
	}

	// Ensure linkTarget is absolute
	if strings.HasPrefix(linkTarget, "..") {
		linkTarget = filepath.Join(filepath.Dir(srcLinkPath), linkTarget)
	}

	// check link does not point outside any directory being copied
	topRelativePath, err := filepath.Rel(topSrcPath, linkTarget)
	if err != nil {
		return gerror.NewFromError(ErrUnexpected, err)
	}
	if strings.HasPrefix(topRelativePath, "..") {
		return gerror.Newf(ErrExternalSymlink,
			"cannot copy symbolic link %q with target %q which points outside the file or directory being copied %q",
			srcLinkPath, linkTarget, topSrcPath)
	}

	linkParent := filepath.Dir(srcLinkPath)
	relativePath, err := filepath.Rel(linkParent, linkTarget)
	if err != nil {
		return gerror.NewFromError(ErrUnexpected, err)
	}
	if glog.V(2) {
		glog.Infof("symbolic link %q has target %q which has path %q relative to %q (directory containing link)",
			srcLinkPath, linkTarget, relativePath, linkParent)
	}
	err = os.Symlink(relativePath, destLinkPath)
	if err != nil {
		return gerror.NewFromError(ErrWritingTargetSymlink, err)
	}
	if glog.V(2) {
		glog.Infof("symbolically linked %q to %q", destLinkPath, relativePath)
	}
	return nil
}

func (f * futils) Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (f *futils) Filemode(path string) (os.FileMode, gerror.Gerror) {
	fi, err := os.Lstat(path)
	if err != nil {
		return os.FileMode(0), gerror.NewFromError(ErrFileNotFound, err)
	}
	return fi.Mode(), nil
}

func (f *futils) sameFile(srcPath string, destPath string) bool {
	srcFi, err := os.Stat(srcPath)
	if err == nil {
		destFi, err := os.Stat(destPath)
		if err == nil {
			return os.SameFile(srcFi, destFi)
		}
	}
	return false
}
