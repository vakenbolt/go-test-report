package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"html/template"
	"io/ioutil"
	"os"
)

type goTestOutputRow struct {
	Time     string
	TestName string `json:"Test"`
	Action   string
	Package  string
	Elapsed  float64
	Output   string
}

type TestStatus struct {
	TestName    string
	Package     string
	ElapsedTime float64
	Output      []string
	Passed      bool
}

type TemplateData struct {
	TestResultGroupIndicatorWidth  string
	TestResultGroupIndicatorHeight string
	TestResults                    []*TestGroupData
	NumOfTestPassed                int
	NumOfTestFailed                int
	NumOfTests                     int
	ReportTitle                    string
	JsCode                         template.JS
}

type TestGroupData struct {
	FailureIndicator string
	TestResults      []*TestStatus
}

type cmdFlags struct {
	titleFlag string
	sizeFlag  string
	groupSize int8
}

func main() {
	stdin := os.Stdin
	flags := cmdFlags{}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the version number of go-test-report",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fmt.Sprintf("go-test-report v%s", "0.9"))
		},
	}
	rootCmd := &cobra.Command{
		Use:  "go-test-report",
		Long: "Captures go test output via stdin and parses it into a single self-contained html file.",
		Run:  func(cmd *cobra.Command, args []string) {
			// start timer
			// -- do stuff
			foobar(stdin)
			// end timer
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
