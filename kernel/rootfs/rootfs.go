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
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ErrorId is used for error ids relating to the RootFS interface.
type ErrorId int

const (
	ErrCreateTempDir ErrorId = iota // a temporary directory for the read-write layer could not be created
	ErrCreateMountDir
	ErrBindMountRoot
	ErrBindMountSubdir
	ErrRootSubdirMissing
	ErrUnmountSubdir
	ErrOverlayTempDir
	ErrOverlayDir
	ErrRemoveMountDir
	ErrUnmountRoot
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

		If Generate fails, it has no side-effects other than possibly
		creating some directories in the read-write base directory.

		The return values are the path of the generated root filesystem
		and an error. The error is `nil` if and only if Generate was
		successful.

		Generate is a replacement for garden's `hook-parent-before-clone.sh`
		script, except that Generate does not copy `wshd` into the
		generated filesystem.
	*/
	Generate(prototype string) (string, gerror.Gerror)

	/*
		Remove a previously generated root filesystem.
	 */
	Remove(root string) gerror.Gerror
}

// ImplErrorId is used for error ids relating to the implementation of this package.
type ImplErrorId int

const (
	ErrNilSyscallFS ImplErrorId = iota // the SyscallFS value was nil
	ErrRwBaseDirMissing                // the read-write base directory was not found
	ErrRwBaseDirIsFile                 // a file was found instead of the read-write base directory
	ErrRwBaseDirNotRw                  // the read-write base directory does not have read and write permissions
)

const tempDirMode os.FileMode = 0777

type rootfs struct {
	sc        syscall.SyscallFS
	f         fileutils.Fileutils
	rwBaseDir string
}

/*
	Creates a new RootFS instance which uses the given SyscallFS interface and the given read-write
	directory as a base for the writable portion of generated root filesystems. The SyscallFS interface value
	must not be nil to ensure that the caller has sufficient privileges to perform mounts, etc.
*/
func NewRootFS(sc syscall.SyscallFS, f fileutils.Fileutils, rwBaseDir string) (RootFS, gerror.Gerror) {
	if sc == nil {
		return nil, gerror.New(ErrNilSyscallFS, "nil SyscallFS")
	}
	fileMode, gerr := f.Filemode(rwBaseDir)
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
	return &rootfs{sc, f, rwBaseDir}, nil
}

func (rfs *rootfs) Generate(prototype string) (root string, gerr gerror.Gerror) {
	if glog.V(1) {
		glog.Infof("Generate(%s)", prototype)
	}
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

	var rwPath string

	if gerr == nil {
		var err error
		rwPath, err = ioutil.TempDir(rfs.rwBaseDir, "tmp-rootfs-")
		defer cleanup(gerror.NewFromError(ErrCreateTempDir, err), func() {
				if e := os.RemoveAll(rwPath); e != nil {
					glog.Warningf("Encountered %q while recovering from %s", e, gerr)
				}
			})
	}

	if gerr == nil {
		var err error
		root, err = ioutil.TempDir(rfs.rwBaseDir, "mnt-")
		defer cleanup(gerror.NewFromError(ErrCreateMountDir, err), func() {
				if e := os.RemoveAll(root); e != nil {
					glog.Warningf("Encountered %q while recovering from %s", e, gerr)
				}
			})
	}

	if gerr == nil {
		err := rfs.sc.BindMountReadOnly(prototype, root)
		if err != nil {
			glog.Errorf("BindMountReadOnly(%s, %s) failed with: %s", prototype, root, err)
		}
		defer cleanup(gerror.NewFromError(ErrBindMountRoot, err), func() {
				if glog.V(1) {
					glog.Infof("unmounting %q", root)
				}
				if e := rfs.sc.Unmount(root); e != nil {
					glog.Warningf("Encountered %q while recovering from %s", e, gerr)
				}
			})
	}

	if gerr == nil {
		gerr = rfs.overlay(root, rwPath)
	}

	return
}

func (rfs *rootfs) overlay(root string, rwPath string) gerror.Gerror {
	if glog.V(2) {
		glog.Infof("overlay(%q, %q)", root, rwPath)
	}
	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}

	// Create a tmp directory so it will end up empty and with the correct permissions regardless of the tmp
	// directory contents and permissions in the prototype root filesystem.
	tmpDir := filepath.Join(rwPath, `tmp`)
	if err := os.Mkdir(tmpDir, tempDirMode); err != nil {
		return gerror.NewFromError(ErrOverlayTempDir, err)
	}

	for i, dir := range dirs {
		if gerr := rfs.overlayDirectory(dir, root, rwPath); gerr != nil {
			for j := i - 1; j >= 0; j-- {
				if cleanupGerr := rfs.unmountOverlayDirectory(dirs[j], root); cleanupGerr != nil {
					glog.Warningf("Encountered %q while recovering from %q", cleanupGerr, gerr)
				}
			}
			return gerr
		}
	}
	return nil
}

func (rfs *rootfs) removeOverlay(root string) gerror.Gerror {
	var firstGerr gerror.Gerror
	dirs := []string{`proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp`}
	for i, _ := range dirs {
		// Reverse the range order.
		dir := dirs[len(dirs) - i - 1]
		if gerr := rfs.unmountOverlayDirectory(dir, root); gerr != nil {
			glog.Errorf("Encountered %s while removing overlay from %s", gerr, root)
			if firstGerr == nil {
				firstGerr = gerr
			}
		}
	}
	return firstGerr
}

func (rfs *rootfs) overlayDirectory(dir string, root string, rwPath string) gerror.Gerror {
	if glog.V(2) {
		glog.Infof("overlayDirectory(%q, %q, %q)", dir, root, rwPath)
	}
	mntPath := filepath.Join(root, dir)
	// check mntPath exists
	if !rfs.f.Exists(mntPath) {
		glog.Errorf("Directory %s not present in root file system (%s)", dir, root)
		return gerror.Newf(ErrRootSubdirMissing, "Directory %s not present in root file system (%s)", dir, root)
	}

	dirPath := filepath.Join(rwPath, dir)

	// Set up read-write directory, copying mount directory contents if there are any.
	if !rfs.f.Exists(dirPath) {
		if err := rfs.f.Copy(dirPath, mntPath); err != nil {
			return gerror.NewFromError(ErrOverlayDir, err)
		}
	}

	if glog.V(2) {
		glog.Infof("BindMountReadWrite(%q, %q)", dirPath, mntPath)
	}
	err := rfs.sc.BindMountReadWrite(dirPath, mntPath)
	if err != nil {
		glog.Errorf("BindMountReadWrite(%s, %s) failed with: %s", dirPath, mntPath, err)
		return gerror.NewFromError(ErrBindMountSubdir, err)
	}
	return nil
}

func (rfs *rootfs) unmountOverlayDirectory(dir string, root string) gerror.Gerror {
	if glog.V(2) {
		glog.Infof("unmountOverlayDirectory(%q, %q)", dir, root)
	}
	mntPath := filepath.Join(root, dir)
	err := rfs.sc.Unmount(mntPath)
	if err != nil {
		return gerror.NewFromError(ErrUnmountSubdir, err)
	}
	return nil
}

func (rfs *rootfs) Remove(root string) gerror.Gerror {
	if glog.V(1) {
		glog.Infof("Remove(%s)", root)
	}
	if gerr := rfs.removeOverlay(root); gerr != nil {
		return gerr
	}
	if err := rfs.sc.Unmount(root); err != nil {
		glog.Errorf("Unmounting %s failed: %s", root, err)
		return gerror.NewFromError(ErrUnmountRoot, err)
	}
	if err := os.RemoveAll(root); err != nil {
		glog.Errorf("Failed to remove %s: %s", root, err)
		return gerror.NewFromError(ErrRemoveMountDir, err)
	}
	return nil
}
