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
Package rootfs encapsulates the container's root filesystem.
*/
package rootfs

import (
	"github.com/cf-guardian/guardian/gerror"
	"github.com/cf-guardian/guardian/kernel/fileutils"
	"github.com/cf-guardian/guardian/kernel/syscall"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// ErrorId is used for error ids relating to the RootFS interface.
type ErrorId int

const (
	ErrCreateTempDir ErrorId = iota
	ErrGetTempDirName
	ErrBindMountRoot
	ErrBindMountSubdir
	ErrOverlayTempDir
	ErrOverlayDir
)

type RootFS interface {
	/*
		Generate produces a usable root filesystem instance from a
		"prototype" root filesystem.

		The input parameter `prototype` is the path of the prototype,
		which may be any filesystem directory. Generate does not modify
		the prototype.

		The resultant filesystem is a collection of read-write
		directories overlaid on the prototype. If an overlay filesystem,
		such as `aufs` or `overlayfs`, is available, it will be used to
		create the result. Otherwise, the result is a mounted filesystem
		consisting of a patchwork quilt of read-write temporary directories
		and the prototype. TODO: decide which of these to support.

		TODO: must this function be called under root?

		If Generate fails, it has no side-effects.

		The return values are the path of the generated root filesystem
		and an error. The error is `nil` if and only if Generate was
		successful.

		Generate is a replacement for garden's `hook-parent-before-clone.sh`
		script, except that Generate does not copy `wshd` into the
		generated filesystem.
	*/
	Generate(prototype string) (string, gerror.Gerror)
}

// ImplErrorId is used for error ids relating to the implementation of this package.
type ImplErrorId int

const (
	ErrRwBaseDirMissing ImplErrorId = iota // the read-write base directory was not found
	ErrRwBaseDirIsFile                     // a file was found instead of the read-write base directory
	ErrRwBaseDirNotRw                      // the read-write base directory does not have read and write permissions
)

const defaultFileMode os.FileMode = 0700
const tempDirMode os.FileMode = 0777

type rootfs struct {
	sc        syscall.Syscall
	rwBaseDir string
}

/*
	Creates a new RootFS instance which uses the given Syscall interface and the given read-write
	directory as a base for the writable portion of generated root filesystems.
*/
func NewRootFS(sc syscall.Syscall, rwBaseDir string) (RootFS, gerror.Gerror) {
	fileMode, gerr := fileutils.Filemode(rwBaseDir)
	if gerr != nil {
		return nil, gerror.NewFromError(ErrRwBaseDirMissing, gerr)
	}
	if !fileMode.IsDir() {
		return nil, gerror.Newf(ErrRwBaseDirIsFile, "File found in place of read-write base directory: %s", rwBaseDir)
	}
	if fileMode.Perm()&0600 != 0600 {
		return nil, gerror.Newf(ErrRwBaseDirNotRw,
			"Read-write base directory does not have read and write permissions: %s has permissions %s",
			rwBaseDir, fileMode.String())
	}
	return &rootfs{sc, rwBaseDir}, nil
}

func (rfs *rootfs) Generate(prototype string) (root string, gerr gerror.Gerror) {
	var err error

	defer func() {
		if gerr != nil {
			root = ""
		}
	}()

	var cleanup = func(capture gerror.Gerror, undo func()) {
		if capture == nil && gerr != nil {
			undo()
		} else {
			gerr = capture
		}
	}

	if err == nil {
		err = os.MkdirAll(rfs.rwBaseDir, defaultFileMode)
		gerr = gerror.NewFromError(ErrCreateTempDir, err)
	}

	var rwPath string

	if err == nil {
		rwPath, err = ioutil.TempDir(rfs.rwBaseDir, "tmp-rootfs")
		var undo = func() {
			if e := os.RemoveAll(rwPath); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, gerr)
			}
		}
		defer cleanup(gerror.NewFromError(ErrGetTempDirName, err), undo)
	}

	if err == nil {
		root, err = ioutil.TempDir(rfs.rwBaseDir, "mnt")
		var undo = func() {
			if e := os.RemoveAll(root); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, gerr)
			}
		}
		defer cleanup(gerror.NewFromError(ErrGetTempDirName, err), undo)
	}

	if err == nil {
		err = rfs.sc.BindMountReadWrite(prototype, root)
		undo := func() {
			if e := rfs.sc.Unmount(root); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, gerr)
			}
		}
		defer cleanup(gerror.NewFromError(ErrBindMountRoot, err), undo)
	}

	if err == nil {
		err = rfs.sc.BindMountReadOnly(prototype, root)
		gerr = gerror.NewFromError(ErrBindMountRoot, err)
	}

	if err == nil {
		gerr = rfs.overlay(root, rwPath)
	}

	return
}

func (rfs *rootfs) overlay(root string, rwPath string) gerror.Gerror {
	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}

	tmpDir := filepath.Join(rwPath, `tmp`)
	if err := os.Mkdir(tmpDir, tempDirMode); err != nil {
		return gerror.NewFromError(ErrOverlayTempDir, err)
	}

	for _, dir := range dirs {
		if gerr := rfs.overlayDirectory(dir, root, rwPath); gerr != nil {
			return gerr
		}
	}
	return nil
}

func (rfs *rootfs) overlayDirectory(dir string, root string, rwPath string) gerror.Gerror {
	dirPath := filepath.Join(rwPath, dir)
	mntPath := filepath.Join(root, dir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if _, err = os.Stat(mntPath); os.IsExist(err) {
			err = fileutils.Copy(dirPath, mntPath)
		} else {
			err = os.MkdirAll(dirPath, tempDirMode)
		}
		if err != nil {
			return gerror.NewFromError(ErrOverlayDir, err)
		}
	}

	err := rfs.sc.BindMountReadWrite(dirPath, mntPath)
	if err != nil {
		return gerror.NewFromError(ErrBindMountSubdir, err)
	}
	return nil
}
