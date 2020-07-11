package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"html/template"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
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
		TestName           string
		Package            string
		ElapsedTime        float64
		Output             []string
		Passed             bool
		TestFileName       string
		TestFunctionDetail TestFunctionFilePos
	}

	TemplateData struct {
		TestResultGroupIndicatorWidth  string
		TestResultGroupIndicatorHeight string
		TestResults                    []*TestGroupData
		NumOfTestPassed                int
		NumOfTestFailed                int
		NumOfTests                     int
		TestDuration                   time.Duration
		ReportTitle                    string
		JsCode                         template.JS
		numOfTestsPerGroup             int
		OutputFilename                 string
		TestExecutionDate              string
	}

	TestGroupData struct {
		FailureIndicator string
		TestResults      []*TestStatus
	}

	cmdFlags struct {
		titleFlag  string
		sizeFlag   string
		groupSize  int
		outputFlag string
		verbose    bool
	}

	GoListJsonModule struct {
		Path string
		Dir  string
		Main bool
	}

	GoListJson struct {
		Dir         string
		ImportPath  string
		Name        string
		GoFiles     []string
		TestGoFiles []string
		Module      GoListJsonModule
	}

	TestFunctionFilePos struct {
		Line int
		Col  int
	}

	TestFileDetail struct {
		FileName            string
		TestFunctionFilePos TestFunctionFilePos
	}

	TestFileDetailsByPackage map[string]map[string]*TestFileDetail
)

func main() {
	rootCmd, _, _ := newRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCommand() (*cobra.Command, *TemplateData, *cmdFlags) {
	flags := &cmdFlags{}
	tmplData := &TemplateData{}
	rootCmd := &cobra.Command{
		Use:  "go-test-report",
		Long: "Captures go test output via stdin and parses it into a single self-contained html file.",
		RunE: func(cmd *cobra.Command, args []string) (e error) {
			startTime := time.Now()
			if err := parseSizeFlag(tmplData, flags); err != nil {
				return err
			}
			tmplData.numOfTestsPerGroup = flags.groupSize
			tmplData.ReportTitle = flags.titleFlag
			tmplData.OutputFilename = flags.outputFlag
			if err := checkIfStdinIsPiped(); err != nil {
				return err
			}
			stdin := os.Stdin
			stdinScanner := bufio.NewScanner(stdin)
			testReportHTMLTemplateFile, _ := os.Create(tmplData.OutputFilename)
			reportFileWriter := bufio.NewWriter(testReportHTMLTemplateFile)
			defer func() {
				_ = stdin.Close()
				if err := reportFileWriter.Flush(); err != nil {
					e = err
				}
				if err := testReportHTMLTemplateFile.Close(); err != nil {
					e = err
				}
			}()
			if err := generateReport(getPackageDetails, stdinScanner, flags, tmplData, reportFileWriter, cmd); err != nil {
				return errors.New(err.Error() + "\n")
			}
			elapsedTime := time.Since(startTime)
			elapsedTimeMsg := []byte(fmt.Sprintf("[go-test-report] finished in %s\n", elapsedTime))
			if _, err := cmd.OutOrStdout().Write(elapsedTimeMsg); err != nil {
				return err
			}
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
	rootCmd.PersistentFlags().StringVarP(&flags.titleFlag,
		"title",
		"t",
		"go-test-report",
		"the title text shown in the test report")
	rootCmd.PersistentFlags().StringVarP(&flags.sizeFlag,
		"size",
		"s",
		"24",
		"the size (in pixels) of the clickable indicator for test result groups")
	rootCmd.PersistentFlags().IntVarP(&flags.groupSize,
		"groupSize",
		"g",
		20,
		"the number of tests per test group indicator")
	rootCmd.PersistentFlags().StringVarP(&flags.outputFlag,
		"output",
		"o",
		"test_report.html",
		"the HTML output file")
	rootCmd.PersistentFlags().BoolVarP(&flags.verbose,
		"verbose",
		"v",
		false,
		"while processing, show the complete output from go test ")

	return rootCmd, tmplData, flags
}

func generateReport(getPackageDetails func(allPackageNames map[string]*types.Nil) (TestFileDetailsByPackage, error),
	stdinScanner *bufio.Scanner, flags *cmdFlags, tmplData *TemplateData, reportFileWriter *bufio.Writer, cmd *cobra.Command) (e error) {
	var err error
	var allTests = map[string]*TestStatus{}
	var allPackageNames = map[string]*types.Nil{}

	// read from stdin and parse "go test" results
	startTestTime := time.Now()
	for stdinScanner.Scan() {
		stdinScanner.Text()
		lineInput := stdinScanner.Bytes()
		if flags.verbose {
			newline := []byte("\n")
			if _, err := cmd.OutOrStdout().Write(append(lineInput, newline[0])); err != nil {
				return err
			}
		}
		goTestOutputRow := &goTestOutputRow{}
		if err := json.Unmarshal(lineInput, goTestOutputRow); err != nil {
			return err
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
			allPackageNames[goTestOutputRow.Package] = nil
			testStatus.Output = append(testStatus.Output, goTestOutputRow.Output)
		}
	}
	elapsedTestTime := time.Since(startTestTime)

	// used to the location of test functions in test go files by package and test function name.
	testFileDetailByPackage, err := getPackageDetails(allPackageNames)
	if err != nil {
		return err
	}

	// read the html template from the generated embedded asset go file
	tpl := template.New("test_report.html.template")
	testReportHtmlTemplateStr, err := hex.DecodeString(testReportHtmlTemplate)
	if err != nil {
		return err
	}
	if tpl, err := tpl.Parse(string(testReportHtmlTemplateStr)); err != nil {
		return err
	} else {
		// read Javascript code from the generated embedded asset go file
		testReportJsCodeStr, err := hex.DecodeString(testReportJsCode)
		if err != nil {
			return err
		}

		tmplData.NumOfTestPassed = 0
		tmplData.NumOfTestFailed = 0
		tmplData.JsCode = template.JS(testReportJsCodeStr)
		tgCounter := 0
		tgId := 0

		// sort the allTests map by test name (this will produce a consistent order when iterating through the map)
		var testNames []string
		for test := range allTests {
			testNames = append(testNames, test)
		}
		sort.Strings(testNames)
		for _, testName := range testNames {
			status := allTests[testName]
			if len(tmplData.TestResults) == tgId {
				tmplData.TestResults = append(tmplData.TestResults, &TestGroupData{})
			}
			// add file info(name and position; line and col) associated with the test function
			testFileInfo := testFileDetailByPackage[status.Package][status.TestName]
			if testFileInfo != nil {
				status.TestFileName = testFileInfo.FileName
				status.TestFunctionDetail = testFileInfo.TestFunctionFilePos
			}
			tmplData.TestResults[tgId].TestResults = append(tmplData.TestResults[tgId].TestResults, status)
			if !status.Passed {
				tmplData.TestResults[tgId].FailureIndicator = "failed"
				tmplData.NumOfTestFailed += 1
			} else {
				tmplData.NumOfTestPassed += 1
			}
			tgCounter += 1
			if tgCounter == tmplData.numOfTestsPerGroup {
				tgCounter = 0
				tgId += 1
			}
		}
		tmplData.NumOfTests = tmplData.NumOfTestPassed + tmplData.NumOfTestFailed
		tmplData.TestDuration = elapsedTestTime.Round(time.Millisecond)
		td := time.Now()
		tmplData.TestExecutionDate = fmt.Sprintf("%s %d, %d %02d:%02d:%02d",
			td.Month(), td.Day(), td.Year(), td.Hour(), td.Minute(), td.Second())
		if err := tpl.Execute(reportFileWriter, tmplData); err != nil {
			return err
		}
	}
	return nil
}

func getPackageDetails(allPackageNames map[string]*types.Nil) (TestFileDetailsByPackage, error) {
	var out bytes.Buffer
	var cmd *exec.Cmd
	testFileDetailByPackage := TestFileDetailsByPackage{}
	stringReader := strings.NewReader("")
	for packageName := range allPackageNames {
		cmd = exec.Command("go", "list", "-json", packageName)
		out.Reset()
		stringReader.Reset("")
		cmd.Stdin = stringReader
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		goListJson := &GoListJson{}
		if err := json.Unmarshal(out.Bytes(), goListJson); err != nil {
			return nil, err
		}
		testFileDetailByPackage[packageName] = map[string]*TestFileDetail{}
		for _, file := range goListJson.TestGoFiles {
			sourceFilePath := fmt.Sprintf("%s/%s", goListJson.Dir, file)
			fileSet := token.NewFileSet()
			f, err := parser.ParseFile(fileSet, sourceFilePath, nil, 0)
			if err != nil {
				return nil, err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					testFileDetail := &TestFileDetail{}
					fileSetPos := fileSet.Position(n.Pos())
					folders := strings.Split(fileSetPos.String(), "/")
					fileNameWithPos := folders[len(folders)-1]
					fileDetails := strings.Split(fileNameWithPos, ":")
					lineNum, _ := strconv.Atoi(fileDetails[1])
					colNum, _ := strconv.Atoi(fileDetails[2])
					testFileDetail.FileName = fileDetails[0]
					testFileDetail.TestFunctionFilePos = TestFunctionFilePos{
						Line: lineNum,
						Col:  colNum,
					}
					testFileDetailByPackage[packageName][x.Name.Name] = testFileDetail
				}
				return true
			})
		}
	}
	return testFileDetailByPackage, nil
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

func checkIfStdinIsPiped() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return nil
	} else {
		return errors.New("ERROR: missing ≪ stdin ≫ pipe")
	}
}
