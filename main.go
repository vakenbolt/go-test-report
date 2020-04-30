package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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

func main() {
	var allTests = map[string]*TestStatus{}

	// read from stdin and parse "go test" results
	stdin := os.Stdin
	defer func() {
		if err := stdin.Close(); err != nil {
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

		templateData := TemplateData{
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
			if len(templateData.TestResults) == tgId {
				templateData.TestResults = append(templateData.TestResults, &TestGroupData{})
			}
			templateData.TestResults[tgId].TestResults = append(templateData.TestResults[tgId].TestResults, status)
			if !status.Passed {
				templateData.TestResults[tgId].FailureIndicator = "failed"
				templateData.NumOfTestFailed += 1
			} else {
				templateData.NumOfTestPassed += 1
			}
			tgCounter += 1
			if tgCounter == numOfTestsPerGroup {
				tgCounter = 0
				tgId += 1
			}
		}
		templateData.NumOfTests = templateData.NumOfTestPassed + templateData.NumOfTestFailed
		err = tpl.Execute(w, templateData)
	}
}
