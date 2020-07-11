package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go/types"
	"io/ioutil"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"version"})
	rootCmdErr := rootCmd.Execute()
	assertions.Nil(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal(fmt.Sprintf("go-test-report v%s\n", version), string(output))
}

func TestTitleFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--title", "Sample Test Report"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("Sample Test Report", tmplData.ReportTitle)
	assertions.NotEmpty(output)
}

func TestTitleFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--title"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --title`)
}

func TestSizeFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, flags := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--size", "24"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("24", flags.sizeFlag)
	assertions.Equal("24px", tmplData.TestResultGroupIndicatorWidth)
	assertions.Equal("24px", tmplData.TestResultGroupIndicatorHeight)
	assertions.NotEmpty(output)
}

func TestSizeFlagWithFullDimensions(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, flags := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--size", "24x16"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("24x16", flags.sizeFlag)
	assertions.Equal("24px", tmplData.TestResultGroupIndicatorWidth)
	assertions.Equal("16px", tmplData.TestResultGroupIndicatorHeight)
	assertions.NotEmpty(output)
}

func TestSizeFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--size"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --size`)
}

func TestGroupSizeFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--groupSize", "32"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal(32, tmplData.numOfTestsPerGroup)
	assertions.NotEmpty(output)
}

func TestGroupSizeFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--groupSize"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --groupSize`)
}

func TestGroupOutputFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--output", "test_file.html"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("test_file.html", tmplData.OutputFilename)
	assertions.NotEmpty(output)
}

func TestGroupOutputFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--output"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --output`)
}

func TestGenerateReport(t *testing.T) {
	assertions := assert.New(t)
	flags := &cmdFlags{}
	data := `{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"go-test-report","Test":"TestFunc1"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc1","Output":"=== RUN   TestFunc1\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc1","Output":"--- PASS: TestFunc1 (1.25s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","Package":"go-test-report","Test":"TestFunc1","Elapsed":1.25}
{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"package2","Test":"TestFunc2"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"package2","Test":"TestFunc2","Output":"=== RUN   TestFunc2\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"package2","Test":"TestFunc2","Output":"--- PASS: TestFunc2 (0.25s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","Package":"package2","Test":"TestFunc2","Elapsed":0.25}
{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"go-test-report","Test":"TestFunc3"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc3","Output":"=== RUN   TestFunc3\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc3","Output":"--- FAIL: TestFunc3 (0.00s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"fail","Package":"go-test-report","Test":"TestFunc3","Elapsed":0}
`
	stdinScanner := bufio.NewScanner(strings.NewReader(data))
	tmplData := &templateData{
		TestResultGroupIndicatorWidth:  "20px",
		TestResultGroupIndicatorHeight: "16px",
		ReportTitle:                    "test-title",
		numOfTestsPerGroup:             2,
		OutputFilename:                 "test-output-report.html",
	}
	cmd := &cobra.Command{}
	writer := bufio.NewWriter(&bytes.Buffer{})
	getPackageDetails := func(allPackageNames map[string]*types.Nil) (testFileDetailsByPackage, error) {
		m := testFileDetailsByPackage{}
		m["go-test-report"] = map[string]*testFileDetail{}
		m["go-test-report"]["TestFunc1"] = &testFileDetail{
			FileName: "sample1.go",
			TestFunctionFilePos: testFunctionFilePos{
				Line: 101,
				Col:  1,
			},
		}
		m["package2"] = map[string]*testFileDetail{}
		m["package2"]["TestFunc2"] = &testFileDetail{
			FileName: "sample2.go",
			TestFunctionFilePos: testFunctionFilePos{
				Line: 784,
				Col:  17,
			},
		}
		return m, nil
	}
	err := generateReport(getPackageDetails, stdinScanner, flags, tmplData, writer, cmd)
	assertions.Nil(err)
	assertions.Equal(2, tmplData.NumOfTestPassed)
	assertions.Equal(1, tmplData.NumOfTestFailed)
	assertions.Equal(3, tmplData.NumOfTests)

	assertions.Equal("TestFunc1", tmplData.TestResults[0].TestResults[0].TestName)
	assertions.Equal("go-test-report", tmplData.TestResults[0].TestResults[0].Package)
	assertions.Equal(true, tmplData.TestResults[0].TestResults[0].Passed)
	assertions.Equal("sample1.go", tmplData.TestResults[0].TestResults[0].TestFileName)
	assertions.Equal(1, tmplData.TestResults[0].TestResults[0].TestFunctionDetail.Col)
	assertions.Equal(101, tmplData.TestResults[0].TestResults[0].TestFunctionDetail.Line)

	assertions.Equal("TestFunc2", tmplData.TestResults[0].TestResults[1].TestName)
	assertions.Equal("package2", tmplData.TestResults[0].TestResults[1].Package)
	assertions.Equal(true, tmplData.TestResults[0].TestResults[1].Passed)
	assertions.Equal("sample2.go", tmplData.TestResults[0].TestResults[1].TestFileName)
	assertions.Equal(17, tmplData.TestResults[0].TestResults[1].TestFunctionDetail.Col)
	assertions.Equal(784, tmplData.TestResults[0].TestResults[1].TestFunctionDetail.Line)

	assertions.Equal("TestFunc3", tmplData.TestResults[1].TestResults[0].TestName)
	assertions.Equal("go-test-report", tmplData.TestResults[1].TestResults[0].Package)
	assertions.Equal(false, tmplData.TestResults[1].TestResults[0].Passed)
	assertions.Empty(tmplData.TestResults[1].TestResults[0].TestFileName)
	assertions.Equal(0, tmplData.TestResults[1].TestResults[0].TestFunctionDetail.Col)
	assertions.Equal(0, tmplData.TestResults[1].TestResults[0].TestFunctionDetail.Line)
}

func TestParseSizeFlagIfValueIsNotInteger(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "x",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `strconv.Atoi: parsing "": invalid syntax`)

}

func TestParseSizeFlagIfWidthValueIsNotInteger(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "Bx27",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `strconv.Atoi: parsing "b": invalid syntax`)
}

func TestParseSizeFlagIfHeightValueIsNotInteger(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "10xA",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `strconv.Atoi: parsing "a": invalid syntax`)
}

func TestParseSizeFlagIfMalformedSize(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "10xx19",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `malformed size value; only one x is allowed if specifying with and height`)
}
