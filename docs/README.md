# go-test-report

[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://shields.io/)
[![version: 0.9](https://img.shields.io/badge/version-0.9-default.svg)](https://shields.io/)
[![version: 0.9](https://img.shields.io/badge/architecture-amd64-darkcyan.svg)](https://shields.io/)
[![version: 0.9](https://img.shields.io/badge/platforms-macos%20|%20windows%20|%20linux-orange.svg)](https://shields.io/)

go-test-report captures `go test` output and parses it into a _single_ self-contained html file. 

## Installation
go-test-report can be installed using [Homebrew](https://brew.sh/)

```shell
$ brew install go-test-report
```

## Usage

To use go-test-report with the default settings. 

```shell script
$ go test -v -json | go-test-report
```

The aforementioned command, outputs an HTML file in the same location. 

```shell
go-test-report.html
```

>Everything needed to display the HTML file correctly is located inside of said file, providing an easy way to store and host the test results.

## Configuration
Additional configuration options are available via command line flags.

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
  -s, --size string     the size of the clickable indicator for test result groups (default "24")
  -t, --title string    the title text shown in the test report (default "go-test-report")
  -v, --verbose         while processing, show the complete output from go test

Use "go-test-report [command] --help" for more information about a command.
```