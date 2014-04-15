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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"github.com/cf-guardian/guardian/kernel/syscall"
	"github.com/cf-guardian/guardian/gerror"
)

const defaultFileMode os.FileMode = 0700
const tempDirMode os.FileMode = 0777

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
func Generate(prototype string, sc syscall.Syscall) (root string, gerr error) {
	var err error

	defer func() {
		if err != nil {
			root = ""
		}
	}()

	var cleanup = func(capture error, undo func()) {
		if capture == nil && err != nil {
			undo()
		} else {
			gerr = capture
		}
	}

	if err == nil {
		err = os.Mkdir("/tmp/guardian", defaultFileMode)

		defer func(capture error) {
			if capture == nil && err != nil {
				e := os.RemoveAll("/tmp/guardian")
				if e != nil {
					log.Printf("Encountered %q while recovering from %q", e, gerr)
				}
			} else {
				gerr = capture
			}
		}(gerror.FromError(err))
	}

	var rwPath string

	if err == nil {
		rwPath, err = ioutil.TempDir("/tmp/guardian", "tmp-rootfs")
		var undo = func() {
			if e := os.RemoveAll(rwPath); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, gerr)
			}
		}
		defer cleanup(gerror.FromError(err), undo)
	}

	if err == nil {
		root, err = ioutil.TempDir("/tmp/guardian", "mnt")
		var undo = func() {
			if e := os.RemoveAll(root); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, gerr)
			}
		}
		defer cleanup(gerror.FromError(err), undo)
	}

	if err == nil {
		err = sc.BindMount(prototype, root, syscall.NO_FLAGS)
		undo := func() {
			if e := sc.Unmount(root); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, gerr)
			}
		}
		defer cleanup(gerror.FromError(err), undo)
	}

	if err == nil {
		err = sc.BindMount(prototype, root, syscall.MS_RDONLY)
		gerr = gerror.FromError(err)
	}

	if err == nil {
		gerr = overlay(root, rwPath)
	}

	return
}

func overlay(root string, rwPath string) error {
	dirs := []string{ `proc`, `dev`, `etc`, `home`, `sbin`, `var`, `tmp` }

	tmpDir := filepath.Join(rwPath, `tmp`)
	if err := os.Mkdir(tmpDir, tempDirMode); err != nil {
		return gerror.FromError(err)
	}

	for _, dir := range dirs {
		if gerr := overlayDirectory(dir, root, rwPath); gerr != nil {
			return gerr
		}
	}
}

func overlayDirectory(dir string, root string, rwPath string) error {
	dirPath := filepath.Join(rwPath, dir)
	mntPath := filepath.Join(root, dir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if _, err := os.Stat(mntPath); os.IsExist(err) {

			// TODO: abstract a recursive copy function out of the following...
			// TODO: need to copy file permissions and attributes
			var walk = func(path string, info os.FileInfo, err error) error {
				if err != nil {
					if info.IsDir() {
						return filepath.SkipDir
					} else {
						return err
					}
				}
				// TODO: mntPath is a prefix of path. Need to add same suffix to dirPath and then call copyFile
				return nil
			}
			err = filepath.Walk(mntPath, filepath.WalkFunc(walk))
		}
	}
	return nil
}
