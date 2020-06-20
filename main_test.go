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
	rootCmd, _ := newRootCommand()
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
	rootCmd, flags := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--title", "Sample Test Report"})
	rootCmdErr := rootCmd.Execute()
	assertions.Nil(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("Sample Test Report", flags.titleFlag)
	assertions.Empty(output)
}

func TestTitleFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _ := newRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--title"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --title`)
}

/*
func Foobar(t *testing.T) {
	assertions := assert.New(t)
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"--title", "foobar"})
	assertions.Nil(rootCmd.Execute())

	foo, err := ioutil.ReadAll(b)
	assertions.Nil(err)
	fmt.Println(err)

	//fmt.Println(output)
	//fmt.Println(string(output))

	fmt.Println(string(foo))
	fmt.Println(flags)

	//fmt.Println(output.Flush())
	//fmt.Println(foobar.Size())
	//fmt.Println(foobar.Flush())
	assertions.True(true)
}
*/