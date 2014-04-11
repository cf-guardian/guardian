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
	"github.com/cf-guardian/guardian/kernel/syscall"
	"github.com/cf-guardian/guardian/gerror"
)

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
func Generate(prototype string, sc syscall.Syscall) (root string, err error) {
        os.MkdirAll("/tmp/guardian", 0700)
	var rwPath string
	rwPath, err = ioutil.TempDir("/tmp/guardian", "tmp-rootfs")
	if err != nil {
		err = gerror.FromError(err)
	} else {
		root, err = ioutil.TempDir("/tmp/guardian", "mnt")
		if err != nil {
			err = gerror.FromError(err)
		} else {
			err = sc.BindMount(prototype, root, syscall.NO_FLAGS)
			if err != nil {
				err = gerror.FromError(err)
			} else {
				err = sc.BindMount(prototype, root, syscall.MS_RDONLY)

				if err != nil {
					err = gerror.FromError(err)
					if e := sc.Unmount(root); e != nil {
						log.Printf("Encountered %q while recovering from %q", e, err)
					}
				}
			}

			if err != nil {
				if e := os.Remove(root); e != nil {
					log.Printf("Encountered %q while recovering from %q", e, err)
				}
			}
		}

		if err != nil {
			if e := os.Remove(rwPath); e != nil {
				log.Printf("Encountered %q while recovering from %q", e, err)
			}
		}
	}

	return root, err
}
