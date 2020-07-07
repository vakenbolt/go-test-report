# go-test-report

[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://shields.io/)
[![version: 0.9](https://img.shields.io/badge/version-0.9-default.svg)](https://shields.io/)
[![version: 0.9](https://img.shields.io/badge/platforms-macos%20|%20windows%20|%20linux-orange.svg)](https://shields.io/)

go-test-report captures `go test` output and parses it into a _single_ self-contained HTML file. 

## Installation

Install the go binary using `go get`

```shell
$ go get -u github.com/vakenbolt/go-test-report/
```

## Usage

To use `go-test-report` with the default settings. 

```shell script
$ go test -json | go-test-report
```

The aforementioned command outputs an HTML file in the same location. 

```shell
go-test-report.html
```

>Everything needed to display the HTML file correctly is located inside of the file, providing an easy way to store and host the test results.

## Configuration
Additional configuration options are available via command-line flags.

```
Usage:
  go-test-report [flags]
  go-test-report [command]

Available Commands:
  help        Help about any command
  version     Prints the version number of go-test-report

Flags:
  -g, --groupSize int   the number of tests per test group indicator (default 10)
  -h, --help            help for go-test-report
  -o, --output string   the HTML output file (default "test_report.html")
  -s, --size string     the size (in pixels) of the clickable indicator for test result groups (default "24")
  -t, --title string    the title text shown in the test report (default "go-test-report")
  -v, --verbose         while processing, show the complete output from go test

Use "go-test-report [command] --help" for more information about a command.
```

The name of the default output file can be changed by using the `-o` or `--output` flag. For example, the following command will change the output to _my-test-report.html_.

```bash
$ go test -json | go-test-report -o my-test-report.html
```


To change the default title shown in the `go-test-report.html` file.

```bash
$ go test -json | go-test-report -t "My Test Report"
```

Use the `-s` or `--size` flag to change the default size of the _group size indicator_. For example, the following command will set the size of the indicator for both the width and height to 48 pixels. 

```bash
$ go test -json | go-test-report -s 48
``` 

Additionally, _both_ the width and height of the _group size indicator_ can be set. For example, the following command will set the size of the indicator to a width of 32 pixels and a height of 16 pixels.

```bash
$ go test -json | go-test-report -g 32x16
```

## Building from source
Before running `go build`, build the "embed_assets" program _inside_ of the `embed_assets` folder. To do so, change the working directory to `%PROJECT_HOME%/embed_assets` and run the following commands:

```bash
$ go build
$ go ./embed_assets
```   

> This will generate an `embedded_assets.go` file in the root directory of the project. This file contains the embedded templates used by `go-test-report`. 

Once the _"embedded_assets.go"_ have been generated, the main binary can be built.

```bash
$ go build
```

To build release binaries (not supported on windows):

```bash
$ ./build_release.sh
```
> Creates a folder in the root project folder called `release_builds` that contains builds for the following platforms:
> - darwin/amd64 (MacOS)
> - linux/amd64
> - windows/amd64
> - windows/i386
