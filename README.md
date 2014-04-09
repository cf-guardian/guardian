# Guardian

Experiments to improve the testability of [garden](https://github.com/pivotal-cf-experimental/garden).

## Development Environment Setup

1. Ensure the following pre-requisites are installed:
    * [git](http://git-scm.com/downloads)
    * [Go](http://golang.org/) 1.2.1 or later: either [download](http://golang.org/doc/install) a specific version or use [gvm](https://github.com/moovweb/gvm).

2. Create a Go [workspace](http://golang.org/doc/code.html#Organization) directory, such as `$HOME/go`, and add the path of this directory to the beginning of `$GOPATH`.

3. Get this repository into your workspace by issuing:
    ```
    $ go get github.com/cf-guardian/guardian
    ```

4. Change directory to `<workspace dir>/src/github.com/cf-guardian/guardian`.

## Testing

Issue:
```
go test
```

If the tests succeed, this should print `PASS`.

## Editing

If your favourite text editor is not sufficient, try [Eclipse](http://www.eclipse.org/downloads/) with the [goclipse plugin](https://github.com/sesteel/goclipse) or [IntelliJ IDEA](http://www.jetbrains.com/idea/) with the [go plugin](https://github.com/go-lang-plugin-org/go-lang-idea-plugin).

## Contributing
[Pull requests](http://help.github.com/send-pull-requests) are welcome; see the [contributor guidelines](CONTRIBUTING.md) for details.

## License
This buildpack is released under version 2.0 of the [Apache License](http://www.apache.org/licenses/LICENSE-2.0).  See the [`LICENSE`](LICENSE) file.
