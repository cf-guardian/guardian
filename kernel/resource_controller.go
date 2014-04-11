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
Package kernel encapsulates the operating system features required
to create a container.
*/
package kernel

// ResourceController provides containment for a specific type of resource.
type ResourceController interface {
	Init(rCtx ResourceContext) error
}

// ResourceContext provides configuration for resource controllers.
type ResourceContext interface {

	// GetRootFS returns the path of the root file system. A root file system is an
	// arbitrary filesystem directory.
	GetRootFS() string
}

type resourceContext struct {
	rootfs string
}

func (rCtx *resourceContext) GetRootFS() string {
	return rCtx.rootfs
}

// CreateResourceContext creates a ResourceContext with the given root file system.
func CreateResourceContext(rootfs string) *resourceContext {
	return &resourceContext{rootfs}
}
