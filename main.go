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
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
		numOfTestsPerGroup             int
		OutputFilename                 string
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
	}
)

func generateTestReport(tmplData *TemplateData, cmd *cobra.Command) error {
	stdin := os.Stdin
	if err := checkIfStdinIsPiped(cmd); err != nil {
		//fmt.Println(err.Error())
		//os.Exit(1)
		return err
	}

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
		return err
	} else {
		testReportHTMLTemplateFile, _ := os.Create(tmplData.OutputFilename)
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
			return err
		}

		//tmplData := TemplateData{
		//	ReportTitle:                    "go-test-report",
		//	TestResultGroupIndicatorWidth:  "24px",
		//	TestResultGroupIndicatorHeight: "24px",
		//	NumOfTestPassed:                0,
		//	NumOfTestFailed:                0,
		//	TestResults:                    []*TestGroupData{},
		//	NumOfTests:                     0,
		//	JsCode:                         template.JS(jsCode),
		//	numOfTestsPerGroup:             20,
		//}
		tmplData.NumOfTestPassed = 0
		tmplData.NumOfTestFailed = 0
		tmplData.JsCode = template.JS(jsCode)
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
			if tgCounter == tmplData.numOfTestsPerGroup {
				tgCounter = 0
				tgId += 1
			}
		}
		tmplData.NumOfTests = tmplData.NumOfTestPassed + tmplData.NumOfTestFailed
		err = tpl.Execute(w, tmplData)
	}
	return nil
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

func checkIfStdinIsPiped(rootCmd *cobra.Command) error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return nil
	} else {
		if err := rootCmd.Help(); err != nil {
			return err
		}
		return errors.New("ERROR: missing ≪ stdin ≫ pipe")
	}
}

func newRootCommand() (*cobra.Command, *TemplateData, *cmdFlags) {
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
			tmplData.numOfTestsPerGroup = flags.groupSize
			tmplData.ReportTitle = flags.titleFlag
			tmplData.OutputFilename = flags.outputFlag
			if err := generateTestReport(tmplData, cmd); err != nil {
				return err
			}
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
	rootCmd.PersistentFlags().StringVarP(&flags.titleFlag,
		"title",
		"t",
		"go-test-report",
		"the title text shown in the test report")
	rootCmd.PersistentFlags().StringVarP(&flags.sizeFlag,
		"size",
		"s",
		"24",
		"the size of the clickable indicator for test result groups")
	rootCmd.PersistentFlags().IntVarP(&flags.groupSize,
		"groupSize",
		"g",
		10,
		"the number of tests per test group indicator")
	rootCmd.PersistentFlags().StringVarP(&flags.outputFlag,
		"output",
		"o",
		"test_report.html",
		"the HTML output file")

	return rootCmd, tmplData, flags
}

func pipeEchoCmdToShellCmd(echoCmd *exec.Cmd, shellCmd *exec.Cmd) {
	var err error
	rPipe, wPipe := io.Pipe()
	echoCmd.Stdout = wPipe
	shellCmd.Stdin = rPipe
	stdoutPipe, err := shellCmd.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	stderrPipe, err := shellCmd.StderrPipe()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	errorCheck(echoCmd.Start())
	errorCheck(shellCmd.Start())
	errorCheck(echoCmd.Wait())
	errorCheck(wPipe.Close())
	scanStdOutErrWithPipeToConsole(&stdoutPipe, &stderrPipe, true)
	errorCheck(shellCmd.Wait())
}

func scanStdOutErrWithPipeToConsole(stdout *io.ReadCloser, stderr *io.ReadCloser, useColor bool) {
	stdoutScanner := bufio.NewScanner(*stdout)
	stdoutScanner.Split(bufio.ScanLines)
	for stdoutScanner.Scan() {
		m := stdoutScanner.Text()
		fmt.Println(m)
	}
	stderrScanner := bufio.NewScanner(*stderr)
	stderrScanner.Split(bufio.ScanLines)
	for stderrScanner.Scan() {
		fmt.Println(stderrScanner.Text())
	}
}

func errorCheck(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

type (
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
)

type FunctionDetail struct {
	FunctionName string
	Line         int
	Col          int
}
type TestFileDetail struct {
	FileName  string
	Functions []FunctionDetail
}

func getPackageDetails(paths []string) error {
	var out bytes.Buffer
	paths = append([]string{"list"}, paths...)
	cmd := exec.Command("go", paths...)
	cmd.Stdin = strings.NewReader("")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	goListOutput := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, packageName := range goListOutput {
		cmd = exec.Command("go", "list", "-json", packageName)
		out.Reset()
		cmd.Stdin = strings.NewReader("")
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		goListJson := &GoListJson{}
		if err := json.Unmarshal(out.Bytes(), goListJson); err != nil {
			return err
		}
		testFileDetails := map[string]TestFileDetail{}
		for _, file := range goListJson.TestGoFiles {
			sourceFilePath := fmt.Sprintf("%s/%s", goListJson.Dir, file)
			fileSet := token.NewFileSet()
			f, err := parser.ParseFile(fileSet, sourceFilePath, nil, 0)
			if err != nil {
				return err
			}
			testFileDetail := TestFileDetail{}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					fileSetPos := fileSet.Position(n.Pos())
					folders := strings.Split(fileSetPos.String(), "/")
					fileNameWithPos := folders[len(folders)-1]
					fileDetails := strings.Split(fileNameWithPos, ":")
					lineNum, _ := strconv.Atoi(fileDetails[1])
					colNum, _ := strconv.Atoi(fileDetails[2])
					testFileDetail.FileName = fileDetails[0]
					testFileDetail.Functions = append(testFileDetail.Functions, FunctionDetail{
						FunctionName: x.Name.Name,
						Line:         lineNum,
						Col:          colNum,
					})
					testFileDetails[sourceFilePath] = testFileDetail
				}
				return true
			})
			o, _ := json.Marshal(testFileDetails)
			fmt.Println(string(o))
		}
	}
	return nil
}

func main() {
	rootCmd, _, _ := newRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	/*
		if err := getPackageDetails([]string{"./..."}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	 */
}