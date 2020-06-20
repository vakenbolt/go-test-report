package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"html/template"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var version = "0.9"

type (
	goTestOutputRow struct {
		Time     string
		TestName string `json:"Test"`
		Action   string
		Package  string
		Elapsed  float64
		Output   string
	}

	TestStatus struct {
		TestName    string
		Package     string
		ElapsedTime float64
		Output      []string
		Passed      bool
	}

	TemplateData struct {
		TestResultGroupIndicatorWidth  string
		TestResultGroupIndicatorHeight string
		TestResults                    []*TestGroupData
		NumOfTestPassed                int
		NumOfTestFailed                int
		NumOfTests                     int
		ReportTitle                    string
		JsCode                         template.JS
	}

	TestGroupData struct {
		FailureIndicator string
		TestResults      []*TestStatus
	}

	cmdFlags struct {
		titleFlag  string
		sizeFlag   string
		groupSize  int8
		outputFlag string
	}
)

func foobar(stdin *os.File) {
	var err error
	var allTests = map[string]*TestStatus{}

	// read from stdin and parse "go test" results

	defer func() {
		if err = stdin.Close(); err != nil {
			panic(err)
		}
	}()
	stdinScanner := bufio.NewScanner(stdin)
	for stdinScanner.Scan() {
		stdinScanner.Text()
		lineInput := stdinScanner.Bytes()
		goTestOutputRow := &goTestOutputRow{}
		if err := json.Unmarshal(lineInput, goTestOutputRow); err != nil {
			fmt.Println(err)
		}
		if goTestOutputRow.TestName != "" {
			var testStatus *TestStatus
			if _, exists := allTests[goTestOutputRow.TestName]; !exists {
				testStatus = &TestStatus{
					TestName: goTestOutputRow.TestName,
					Package:  goTestOutputRow.Package,
					Output:   []string{},
				}
				allTests[goTestOutputRow.TestName] = testStatus
			} else {
				testStatus = allTests[goTestOutputRow.TestName]
			}
			if goTestOutputRow.Action == "pass" || goTestOutputRow.Action == "fail" {
				if goTestOutputRow.Action == "pass" {
					testStatus.Passed = true
				}
				testStatus.ElapsedTime = goTestOutputRow.Elapsed
			}
			testStatus.Output = append(testStatus.Output, goTestOutputRow.Output)
		}
	}
	if tpl, err := template.ParseFiles("test_report.html.template"); err != nil {
		panic(err)
	} else {
		testReportHTMLTemplateFile, _ := os.Create("test_report.html")
		w := bufio.NewWriter(testReportHTMLTemplateFile)
		defer func() {
			if err := w.Flush(); err != nil {
				fmt.Println(err)
			}
			if err := testReportHTMLTemplateFile.Close(); err != nil {
				panic(err)
			}
		}()

		// read Javascript test code
		jsCode, err := ioutil.ReadFile("test_report.js")
		if err != nil {
			panic(err)
		}

		tmplData := TemplateData{
			ReportTitle:                    "go-test-report",
			TestResultGroupIndicatorWidth:  "24px",
			TestResultGroupIndicatorHeight: "24px",
			NumOfTestPassed:                0,
			NumOfTestFailed:                0,
			TestResults:                    []*TestGroupData{},
			NumOfTests:                     0,
			JsCode:                         template.JS(jsCode),
		}
		numOfTestsPerGroup := 20
		tgCounter := 0
		tgId := 0

		for _, status := range allTests {
			if len(tmplData.TestResults) == tgId {
				tmplData.TestResults = append(tmplData.TestResults, &TestGroupData{})
			}
			tmplData.TestResults[tgId].TestResults = append(tmplData.TestResults[tgId].TestResults, status)
			if !status.Passed {
				tmplData.TestResults[tgId].FailureIndicator = "failed"
				tmplData.NumOfTestFailed += 1
			} else {
				tmplData.NumOfTestPassed += 1
			}
			tgCounter += 1
			if tgCounter == numOfTestsPerGroup {
				tgCounter = 0
				tgId += 1
			}
		}
		tmplData.NumOfTests = tmplData.NumOfTestPassed + tmplData.NumOfTestFailed
		err = tpl.Execute(w, tmplData)
	}
}

func parseSizeFlag(tmplData *TemplateData, flags *cmdFlags) error {
	flags.sizeFlag = strings.ToLower(flags.sizeFlag)
	if !strings.Contains(flags.sizeFlag, "x") {
		if val, err := strconv.Atoi(flags.sizeFlag); err != nil {
			return err
		} else {
			tmplData.TestResultGroupIndicatorWidth = fmt.Sprintf("%dpx", val)
			tmplData.TestResultGroupIndicatorHeight = fmt.Sprintf("%dpx", val)
			return nil
		}
	}
	if strings.Count(flags.sizeFlag, "x") > 1 {
		return errors.New(`malformed size value; only one x is allowed if specifying with and height`)
	} else {
		a := strings.Split(flags.sizeFlag, "x")
		if val, err := strconv.Atoi(a[0]); err != nil {
			return err
		} else {
			tmplData.TestResultGroupIndicatorWidth = fmt.Sprintf("%dpx", val)

		}
		if val, err := strconv.Atoi(a[1]); err != nil {
			return err
		} else {
			tmplData.TestResultGroupIndicatorHeight = fmt.Sprintf("%dpx", val)
		}
		return nil
	}
}

func newRootCommand() (*cobra.Command, *cmdFlags, *TemplateData) {
	flags := &cmdFlags{}
	tmplData := &TemplateData{}
	rootCmd := &cobra.Command{
		Use:  "go-test-report",
		Long: "Captures go test output via stdin and parses it into a single self-contained html file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// start timer
			// -- do stuff
			if err := parseSizeFlag(tmplData, flags); err != nil {
				return err
			}
			//foobar(stdin)
			// end timer
			return nil
		},
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the version number of go-test-report",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg := fmt.Sprintf("go-test-report v%s", version)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return err
			}
			return nil
		},
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentFlags().StringVar(&flags.titleFlag,
		"title",
		"go-test-report",
		"the title text shown in the test report")
	rootCmd.PersistentFlags().StringVar(&flags.sizeFlag,
		"size",
		"24",
		"the size of the clickable indicator for test result groups")
	rootCmd.PersistentFlags().Int8Var(&flags.groupSize,
		"groupSize",
		0,
		"the number of tests per test group")
	rootCmd.PersistentFlags().StringVarP(&flags.outputFlag,
		"output",
		"o",
		"test_report.html",
		"the HTML output file")

	return rootCmd, flags, tmplData
}

func main() {
	rootCmd, _, _ := newRootCommand()
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		fmt.Println("data is being piped to stdin")
	} else {
		if err := rootCmd.Help(); err != nil {
			panic(err)
		}
		fmt.Println("ERROR: missing ≪ stdin ≫ pipe")
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
