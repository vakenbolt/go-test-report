package main

import (
	"encoding/hex"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"
)

type templateData struct {
	TestResultGroupIndicatorWidth  string
	TestResultGroupIndicatorHeight string
	TestResults                    []*testGroupData
	NumOfTestPassed                int
	NumOfTestFailed                int
	NumOfTestSkipped               int
	NumOfTests                     int
	TestDuration                   time.Duration
	ReportTitle                    string
	JsCode                         template.JS
	numOfTestsPerGroup             int
	OutputFilename                 string
	TestExecutionDate              string
}

type testGroupData struct {
	FailureIndicator string
	SkippedIndicator string
	TestResults      []*testStatus
}

func (tp *templateData) initReportHTML() (*template.Template, error) {
	// read the html template from the generated embedded asset go file
	tpl := template.New("test_report.html.template")
	testReportHTMLTemplateStr, err := hex.DecodeString(testReportHTMLTemplate)
	if err != nil {
		return nil, err
	}
	tpl, err = tpl.Parse(string(testReportHTMLTemplateStr))
	if err != nil {
		return nil, err
	}
	// read Javascript code from the generated embedded asset go file
	testReportJsCodeStr, err := hex.DecodeString(testReportJsCode)
	if err != nil {
		return nil, err
	}
	tp.JsCode = template.JS(testReportJsCodeStr)
	return tpl, nil
}

func (tp *templateData) readDataFromHTML(fi *os.File) (err error) {
	doc, _ := goquery.NewDocumentFromReader(fi)
	// readHeader
	doc.Find(".pageHeader .testStats .total strong").Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			a, err := strconv.Atoi(selection.Text())
			if err != nil {
				return
			}
			tp.NumOfTests = a
		} else if i == 1 {
			dua, err := time.ParseDuration(selection.Text())
			if err != nil {
				return
			}
			tp.TestDuration = dua
		}
	})

	passText := doc.Find(".pageHeader .testStats .passed strong").Text()
	passNum, err := strconv.Atoi(passText)
	if err != nil {
		return
	}
	tp.NumOfTestPassed = passNum

	skippedText := doc.Find(".pageHeader .testStats .skipped strong").Text()
	skippedNum, err := strconv.Atoi(skippedText)
	if err != nil {
		return
	}
	tp.NumOfTestSkipped = skippedNum

	failedText := doc.Find(".pageHeader .testStats .failed strong").Text()
	failedNum, err := strconv.Atoi(failedText)
	if err != nil {
		return
	}
	tp.NumOfTestFailed = failedNum

	strs := strings.SplitAfter(doc.Find("script").Text(), "const data = ")
	strs = strings.SplitAfter(strs[1], "}}]}]")
	var tempResult []*testGroupData
	err = json.Unmarshal([]byte(strs[0]), &tempResult)
	if err != nil {
		return
	}
	tp.TestResults = tempResult
	return
}
