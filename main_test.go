package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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
