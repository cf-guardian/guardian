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

/*
	Copy copies a source file to a destination file. File contents are copied. File mode and permissions
	(as described in http://golang.org/pkg/os/#FileMode) are copied.

	Directories are copied, along with their contents.

	Copying a file or directory to itself succeeds but does not modify the filesystem.

	Symbolic links are not followed and are copied provided they refer to a file or directory being copied
	(otherwise a non-nil error is returned). The only exception is copying a symbolic link to itself, which
	always succeeds.
*/
func Copy(destPath string, srcPath string) gerror.Gerror {
	glog.Infof("Copy(%s, %s)", destPath, srcPath)
	return doCopy(destPath, srcPath, srcPath)
}

func doCopy(destPath string, srcPath string, topSrcPath string) gerror.Gerror {
	if sameFile(srcPath, destPath) {
		return nil
	}
	srcMode, gerr := Filemode(srcPath)
	if gerr != nil {
		return gerr
	}

	if srcMode&os.ModeSymlink == os.ModeSymlink {
		return copySymlink(destPath, srcPath, topSrcPath)
	} else if srcMode.IsDir() {
		return copyDir(destPath, srcPath, topSrcPath)
	} else {
		return copyFile(destPath, srcPath)
	}
}

func copyDir(destination string, source string, topSource string) gerror.Gerror {
	glog.Infof("copyDir(%s, %s)", destination, source)
	finalDestination, gerr := finalDestinationDir(destination, source)
	if gerr != nil {
		return gerr
	}
	src, err := os.Open(source)
	if err != nil {
		return gerror.NewFromError(ErrOpeningSourceDir, err)
	}

	names, err := src.Readdirnames(-1)
	if err != nil {
		return gerror.NewFromError(ErrCannotListSourceDir, err)
	}

	for _, name := range names {
		glog.Infof("copying %s from %s to %s", name, source, finalDestination)
		gerr = doCopy(filepath.Join(finalDestination, name), filepath.Join(source, name), topSource)
		if gerr != nil {
			return gerr
		}
	}

	return nil
}

/*
	Determine the final destination directory and return an opened file referring to it.
*/
func finalDestinationDir(destination string, source string) (finalDestination string, gerr gerror.Gerror) {
	defer func() {
		glog.Infof("openFinalDestinationDir(%s, %s) returning (%v, %v)", destination, source, finalDestination, gerr)
	}()
	sourceMode, gerr := Filemode(source)
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

func copyFile(destination string, source string) gerror.Gerror {
	glog.Infof("copyFile(%s, %s)", destination, source)
	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		return gerror.NewFromError(ErrOpeningSourceFile, err)
	}
	defer sourceFile.Close()

	mode, gerr := Filemode(source)
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

func copySymlink(destLinkPath string, srcLinkPath string, topSrcPath string) gerror.Gerror {
	glog.Infof("copySymLink(%s, %s, %s)", destLinkPath, srcLinkPath, topSrcPath)
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
			"cannot copy symbolic link %s with target %s which points outside the file or directory being copied %s",
			srcLinkPath, linkTarget, topSrcPath)
	}

	linkParent := filepath.Dir(srcLinkPath)
	relativePath, err := filepath.Rel(linkParent, linkTarget)
	if err != nil {
		return gerror.NewFromError(ErrUnexpected, err)
	}
	glog.Infof("symbolic link %s has target %s which has path %s relative to %s (directory containing link)",
		srcLinkPath, linkTarget, relativePath, linkParent)
	err = os.Symlink(relativePath, destLinkPath)
	if err != nil {
		return gerror.NewFromError(ErrWritingTargetSymlink, err)
	}
	glog.Infof("symbolically linked %s to %s", destLinkPath, relativePath)
	return nil
}

/*
	Returns the os.FileMode of the file with the given path. If the file does not exist, return an error with tag
	ErrFileNotFound.
*/
func Filemode(path string) (os.FileMode, gerror.Gerror) {
	fi, err := os.Lstat(path)
	if err != nil {
		return os.FileMode(0), gerror.NewFromError(ErrFileNotFound, err)
	}
	return fi.Mode(), nil
}

func sameFile(srcPath string, destPath string) bool {
	srcFi, err := os.Stat(srcPath)
	if err == nil {
		destFi, err := os.Stat(destPath)
		if err == nil {
			return os.SameFile(srcFi, destFi)
		}
	}
	return false
}
