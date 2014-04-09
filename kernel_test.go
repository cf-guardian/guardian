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
package guardian

import (
	"github.com/cf-guardian/guardian/kernel"
	"testing"
)

type testResourceController struct{ a int }

func (testResourceController) Init(rCtx kernel.ResourceContext) error {
	return nil
}

func run(res kernel.ResourceController) {
	rCtx := kernel.CreateResourceContext("/")
	res.Init(rCtx)
}

func TestRC(t *testing.T) {
	var x testResourceController
	run(x)
}
