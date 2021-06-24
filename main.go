package main

import (
	"bufio"
	"bytes"
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

type (
	goTestOutputRow struct {
		Time     string
		TestName string `json:"Test"`
		Action   string
		Package  string
		Elapsed  float64
		Output   string
	}

	testStatus struct {
		TestName           string
		Package            string
		ElapsedTime        float64
		Output             []string
		Passed             bool
		Skipped            bool
		TestFileName       string
		TestFunctionDetail testFunctionFilePos
	}

	cmdFlags struct {
		titleFlag     string
		sizeFlag      string
		groupSize     int
		outputFlag    string
		verbose       bool
		append        bool
		targetPackage string
	}

	goListJSONModule struct {
		Path string
		Dir  string
		Main bool
	}

	goListJSON struct {
		Dir         string
		ImportPath  string
		Name        string
		GoFiles     []string
		TestGoFiles []string
		Module      goListJSONModule
	}

	testFunctionFilePos struct {
		Line int
		Col  int
	}

	testFileDetail struct {
		FileName            string
		TestFunctionFilePos testFunctionFilePos
	}

	testFileDetailsByPackage map[string]map[string]*testFileDetail
)

func main() {
	rootCmd, _, _ := initRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initRootCommand() (*cobra.Command, *templateData, *cmdFlags) {
	flags := &cmdFlags{}
	tmplData := &templateData{}
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
			var (
				testReportHTMLTemplateFile *os.File
				tpl                        *template.Template
			)
			tpl, err := tmplData.initReportHTML()
			if err != nil {
				return err
			}
			if flags.append && FileExist(tmplData.OutputFilename) {
				testReportHTMLTemplateFile, _ = os.OpenFile(tmplData.OutputFilename, os.O_RDWR, 0)
				err := tmplData.readDataFromHTML(testReportHTMLTemplateFile)
				if err != nil {
					return err
				}
			}
			testReportHTMLTemplateFile, _ = os.Create(tmplData.OutputFilename)
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
			startTestTime := time.Now()
			allPackageNames, allTests, err := readTestDataFromStdIn(stdinScanner, flags, cmd)
			if err != nil {
				return errors.New(err.Error() + "\n")
			}
			elapsedTestTime := time.Since(startTestTime)
			// used to the location of test functions in test go files by package and test function name.
			testFileDetailByPackage, err := getPackageDetails(allPackageNames, flags.targetPackage)
			if err != nil {
				return err
			}
			err = generateReport(tpl, tmplData, allTests, testFileDetailByPackage, elapsedTestTime, reportFileWriter)
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
		100,
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
	rootCmd.PersistentFlags().BoolVarP(&flags.append,
		"append",
		"a",
		false,
		"append to the HTML output file")
	rootCmd.PersistentFlags().StringVarP(&flags.targetPackage,
		"package",
		"p",
		"",
		"if run with a test file, point out the target package")

	return rootCmd, tmplData, flags
}

func readTestDataFromStdIn(stdinScanner *bufio.Scanner, flags *cmdFlags, cmd *cobra.Command) (allPackageNames map[string]*types.Nil, allTests map[string]*testStatus, e error) {
	allTests = map[string]*testStatus{}
	allPackageNames = map[string]*types.Nil{}

	// read from stdin and parse "go test" results
	for stdinScanner.Scan() {
		lineInput := stdinScanner.Bytes()
		if flags.verbose {
			newline := []byte("\n")
			if _, err := cmd.OutOrStdout().Write(append(lineInput, newline[0])); err != nil {
				return nil, nil, err
			}
		}
		goTestOutputRow := &goTestOutputRow{}
		if err := json.Unmarshal(lineInput, goTestOutputRow); err != nil {
			continue
		}
		if goTestOutputRow.TestName != "" {
			var status *testStatus
			key := goTestOutputRow.Package + "." + goTestOutputRow.TestName
			if _, exists := allTests[key]; !exists {
				status = &testStatus{
					TestName: goTestOutputRow.TestName,
					Package:  goTestOutputRow.Package,
					Output:   []string{},
				}
				allTests[key] = status
			} else {
				status = allTests[key]
			}
			if goTestOutputRow.Action == "pass" || goTestOutputRow.Action == "fail" || goTestOutputRow.Action == "skip" {
				if goTestOutputRow.Action == "pass" {
					status.Passed = true
				}
				if goTestOutputRow.Action == "skip" {
					status.Skipped = true
				}
				status.ElapsedTime = goTestOutputRow.Elapsed
			}
			allPackageNames[goTestOutputRow.Package] = nil
			if strings.Contains(goTestOutputRow.Output, "--- PASS:") {
				goTestOutputRow.Output = strings.TrimSpace(goTestOutputRow.Output)
			}
			status.Output = append(status.Output, goTestOutputRow.Output)
		}
	}
	return allPackageNames, allTests, nil
}

func getPackageDetails(allPackageNames map[string]*types.Nil, defaultPackage string) (testFileDetailsByPackage, error) {
	var out bytes.Buffer
	var cmd *exec.Cmd
	testFileDetailByPackage := testFileDetailsByPackage{}
	stringReader := strings.NewReader("")
	for packageName := range allPackageNames {
		if packageName == "command-line-arguments" {
			cmd = exec.Command("go", "list", "-json", defaultPackage)
		} else {
			cmd = exec.Command("go", "list", "-json", packageName)
		}
		out.Reset()
		stringReader.Reset("")
		cmd.Stdin = stringReader
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		goListJSON := &goListJSON{}
		if err := json.Unmarshal(out.Bytes(), goListJSON); err != nil {
			return nil, err
		}
		testFileDetailByPackage[packageName] = map[string]*testFileDetail{}
		for _, file := range goListJSON.TestGoFiles {
			sourceFilePath := fmt.Sprintf("%s/%s", goListJSON.Dir, file)
			fileSet := token.NewFileSet()
			f, err := parser.ParseFile(fileSet, sourceFilePath, nil, 0)
			if err != nil {
				return nil, err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					testFileDetail := &testFileDetail{}
					fileSetPos := fileSet.Position(n.Pos())
					folders := strings.Split(fileSetPos.String(), "/")
					fileNameWithPos := folders[len(folders)-1]
					fileDetails := strings.Split(fileNameWithPos, ":")
					lineNum, _ := strconv.Atoi(fileDetails[1])
					colNum, _ := strconv.Atoi(fileDetails[2])
					testFileDetail.FileName = fileDetails[0]
					testFileDetail.TestFunctionFilePos = testFunctionFilePos{
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

type testRef struct {
	key  string
	name string
}
type byName []testRef

func (t byName) Len() int {
	return len(t)
}
func (t byName) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t byName) Less(i, j int) bool {
	return t[i].name < t[j].name
}

func generateReport(tpl *template.Template, tmplData *templateData, allTests map[string]*testStatus, testFileDetailByPackage testFileDetailsByPackage, elapsedTestTime time.Duration, reportFileWriter *bufio.Writer) error {
	tgCounter := 0
	tgID := len(tmplData.TestResults)

	// sort the allTests map by test name (this will produce a consistent order when iterating through the map)
	var tests []testRef
	for test, status := range allTests {
		tests = append(tests, testRef{test, status.TestName})
	}
	sort.Sort(byName(tests))
	for _, test := range tests {
		status := allTests[test.key]
		if len(tmplData.TestResults) == tgID {
			tmplData.TestResults = append(tmplData.TestResults, &testGroupData{})
		}
		// add file info(name and position; line and col) associated with the test function
		testFileInfo := testFileDetailByPackage[status.Package][status.TestName]
		if testFileInfo != nil {
			status.TestFileName = testFileInfo.FileName
			status.TestFunctionDetail = testFileInfo.TestFunctionFilePos
		}
		tmplData.TestResults[tgID].TestResults = append(tmplData.TestResults[tgID].TestResults, status)
		if !status.Passed {
			if !status.Skipped {
				tmplData.TestResults[tgID].FailureIndicator = "failed"
				tmplData.NumOfTestFailed++
			} else {
				tmplData.TestResults[tgID].SkippedIndicator = "skipped"
				tmplData.NumOfTestSkipped++
			}
		} else {
			tmplData.NumOfTestPassed++
		}
		tgCounter++
		if tgCounter == tmplData.numOfTestsPerGroup {
			tgCounter = 0
			tgID++
		}
	}
	tmplData.NumOfTests = tmplData.NumOfTestPassed + tmplData.NumOfTestFailed + tmplData.NumOfTestSkipped
	tmplData.TestDuration += elapsedTestTime.Round(time.Millisecond)
	td := time.Now()
	tmplData.TestExecutionDate = fmt.Sprintf("%s %d, %d %02d:%02d:%02d",
		td.Month(), td.Day(), td.Year(), td.Hour(), td.Minute(), td.Second())
	if err := tpl.Execute(reportFileWriter, tmplData); err != nil {
		return err
	}
	return nil
}

func parseSizeFlag(tmplData *templateData, flags *cmdFlags) error {
	flags.sizeFlag = strings.ToLower(flags.sizeFlag)
	if !strings.Contains(flags.sizeFlag, "x") {
		val, err := strconv.Atoi(flags.sizeFlag)
		if err != nil {
			return err
		}
		tmplData.TestResultGroupIndicatorWidth = fmt.Sprintf("%dpx", val)
		tmplData.TestResultGroupIndicatorHeight = fmt.Sprintf("%dpx", val)
		return nil
	}
	if strings.Count(flags.sizeFlag, "x") > 1 {
		return errors.New(`malformed size value; only one x is allowed if specifying with and height`)
	}
	a := strings.Split(flags.sizeFlag, "x")
	valW, err := strconv.Atoi(a[0])
	if err != nil {
		return err
	}
	tmplData.TestResultGroupIndicatorWidth = fmt.Sprintf("%dpx", valW)
	valH, err := strconv.Atoi(a[1])
	if err != nil {
		return err
	}
	tmplData.TestResultGroupIndicatorHeight = fmt.Sprintf("%dpx", valH)
	return nil
}

func checkIfStdinIsPiped() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return nil
	}
	return errors.New("ERROR: missing ≪ stdin ≫ pipe")
}
