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
The container package is an abstract interface for running commands
in containers.
*/
package container

type InputStream chan string

type OutputStream chan string

type ExitStatus chan int

/*
	A Container is a function which runs a given command with a given input stream and returns an output
	stream, an error stream, an exit status, and an error. If the command cannot be run, a non-nil error
	is returned along with nil for the other return values. If the command can be run, a nil error is
	returned and the command runs asynchronously to the caller. The caller may write to the input stream
	and read from the output and error streams. The exit status is available for reading by the caller once
	the command has returned. The value of the exit status is the return code of the command.
*/
type Container func(command string, input InputStream) (output OutputStream, err OutputStream, status ExitStatus, err error)
