package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
)

var htmlTemplate []byte
var jsCode []byte

func init() {
	htmlTemplate, _ = ioutil.ReadFile("test_report.html.template")
	jsCode, _ =  ioutil.ReadFile("test_report.js")
}

func main() {
	outputFile1, _ := os.Create("../embedded_assets.go")
	writer := bufio.NewWriter(outputFile1)
	defer func() {
		if err := writer.Flush(); err != nil {
			panic(err)
		}
		if err := outputFile1.Close(); err != nil {
			panic(err)
		}
	}()
	dst := make([]byte, hex.EncodedLen(len(htmlTemplate)))
	hex.Encode(dst, htmlTemplate)
	_, _ = writer.WriteString(fmt.Sprintf("package main\n\nvar testReportHtmlTemplateSize = %d\nvar testReportHtmlTemplate = `%s`", len(htmlTemplate), string(dst)))
	dst = make([]byte, hex.EncodedLen(len(jsCode)))
	hex.Encode(dst, jsCode)
	_, _ = writer.WriteString(fmt.Sprintf("\n\nvar testReportJsCodeSize = %d\nvar testReportJsCode = `%s`", len(jsCode), string(dst)))
}

func TestReportHtmlTemplate() ([]byte, error) {
	dst := make([]byte, hex.DecodedLen(len(htmlTemplate)))
	if _, err := hex.Decode(dst, htmlTemplate); err != nil {
		return nil, err
	} else {
		return dst, err
	}
}

func TestReportJsCode() ([]byte, error) {
	dst := make([]byte, hex.DecodedLen(len(jsCode)))
	if _, err := hex.Decode(dst, jsCode); err != nil {
		return nil, err
	} else {
		return dst, err
	}
}