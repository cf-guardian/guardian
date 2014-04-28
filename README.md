[![GoDoc](https://godoc.org/github.com/cf-guardian/guardian/kernel?status.png)](https://godoc.org/github.com/cf-guardian/guardian/kernel)

# Guardian

Experiments to improve the testability of [warden](https://github.com/cloudfoundry-incubator/warden-linux).

## Objectives

1. Understandability: the code should be clearly structured and documented so that a newcomer can understand the rationale and be able to propose changes.
1. Robustness: the code should function correctly or fail with meaningful diagnostics.
1. Maintainability: it should be straightforward to fix bugs and add new features.
1. Testability: the runtime code should be thoroughly exercised by the tests.
1. Portability: it should be straightforward to port the code to other Linux distributions.

These objectives will be achieved through the following practices:

1. Construct separately testable components with documented interfaces.
1. Test each component including error paths.
1. Keep runtime and test code separate.
1. Fail rather than degrade function.
1. Operating system dependencies should be isolated and carefully managed to simplify porting.
1. Use pure Go for maintainability. Avoid scripting (even in Go) and C code.
1. Instrument code to capture failure diagnostics including stack traces.

## Documentation

Documentation is available at [godoc.org](http://godoc.org/github.com/cf-guardian/guardian).

If this hasn't been refreshed for a while, feel free to click the "Refresh now" link.

## Diagnostics

### Errors

Errors returned from Guardian functions and methods include stack traces so that the point of failure can easily be determined. The [gerror](gerror) package is used to construct errors.

### Logging

Logging is performed using the [glog](https://github.com/golang/glog) package (an external dependency). Logs may be directed to standard error by setting the flag `logtostderr` to `true` on the go invocation, as in this example:

````
go test -logtostderr=true
````
See the [glog documentation](http://godoc.org/github.com/golang/glog) for further information.


## Development Environment Setup

1. Ensure the following pre-requisites are installed:
    * [git](http://git-scm.com/downloads)
    * [Go](http://golang.org/) 1.2.1 or later:

        - either [download](http://golang.org/doc/install) a specific version
        - or use [gvm](https://github.com/moovweb/gvm)
        - or even `port install go` with [MacPorts](http://www.macports.org/).

2. Create a Go [workspace](http://golang.org/doc/code.html#Organization) directory, such as `$HOME/go`, and add the path of this directory to the
beginning of a new environment variable called `GOPATH`. You might want to put this last step in your profile.
    ```
    $ mkdir $HOME/go
    $ export GOPATH=$HOME/go
    ```

3. Get this repository into your workspace (`src` directory) by issuing:
    ```
    $ go get github.com/cf-guardian/guardian
    ```

4. Change directory to `<workspace dir>/src/github.com/cf-guardian/guardian`.

5. Install the [pre-commit hook](https://github.com/jbrukh/git-gofmt) as follows:
    ```
    cd .git/hooks
    ln -s ../../development/pre-commit-hook/pre-commit .
    ```

    After installing the hook, if you need to skip reformatting for a particular commit, use `git commit --no-verify`.

## Testing

Issue:
```
go test
```

If the tests succeed, this should print `PASS`.

## Editing

If your favourite text editor is not sufficient, try [Eclipse](http://www.eclipse.org/downloads/) with the [goclipse plugin](https://github.com/sesteel/goclipse) or [IntelliJ IDEA](http://www.jetbrains.com/idea/) with the [go plugin](https://github.com/go-lang-plugin-org/go-lang-idea-plugin).

Source code is formatted according to standard Go conventions. To re-format the code, issue:
```
go fmt ./...
```

To reformat code before committing it to git, install the pre-commit hook as described above.


Also, you can [lint](http://go-lint.appspot.com/github.com/cf-guardian/guardian) the code if you like.

## Contributing
[Pull requests](http://help.github.com/send-pull-requests) are welcome; see the [contributor guidelines](CONTRIBUTING.md) for details.

## License
This buildpack is released under version 2.0 of the [Apache License](http://www.apache.org/licenses/LICENSE-2.0).  See the [`LICENSE`](LICENSE) file.
