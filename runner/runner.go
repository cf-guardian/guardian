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
	The runner package builds a container and runs a single command
	in the container. The container is created from a list of
	resource controllers each of which virtualises a specific type of
	resource so that the command runs in an isolated environment with
	respect to resources of that type.
*/
package runner

import (
	"github.com/cf-guardian/guardian/container"
	"github.com/cf-guardian/guardian/kernel"
)

func BuildContainer(rCtx kernel.ResourceContext, rcs []kernel.ResourceController) container.Container {

	return nil
}
