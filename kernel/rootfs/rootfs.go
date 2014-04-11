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
Package rootfs package encapsulates the container's root filesystem.
*/
package rootfs

import (
	//"path/filepath"
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
	and the prototype.
	
	If Generate fails, it has no side-effects.
	
	The return values are the path of the generated root filesystem
	and an error. The error is `nil` if and only if Generate was
	successful.

	Generate is a replacement for garden's `hook-parent-before-clone.sh`
	script, except that Generate does not copy `wshd` into the
	generated filesystem.
*/
func Generate(prototype string) (root string, err error) {
	return "", nil
}
