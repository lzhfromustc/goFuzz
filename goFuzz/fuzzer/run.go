package fuzzer

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type RunTask struct {
	id    string
	stage FuzzStage
	input *Input
	entry *FuzzQueryEntry
}

func getInputFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "input"))
}

func getOutputFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "stdout"))
}

func getErrFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "stderr"))
}

func getRecordFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "record"))
}

func NewRunTask(input *Input, stage FuzzStage, idx int, entry *FuzzQueryEntry) *RunTask {
	task := &RunTask{
		input: input,
		entry: entry,
		stage: stage,
		id:    fmt.Sprintf("%s-%s-%d", input.TestName, stage, idx),
	}
	return task
}

func Run(task *RunTask) (*RunResult, error) {
	var err error
	input := task.input
	if input.TestName == "Empty" || input.TestName == "" {
		return nil, fmt.Errorf("the Run command in the fuzzer receive an input without input.TestName")
	}

	// Setting up related file paths

	runOutputDir := path.Join(OutputDir, task.id)
	err = createDir(runOutputDir)
	if err != nil {
		return nil, err
	}

	gfInputFp, err := getInputFilePath(runOutputDir)
	if err != nil {
		return nil, err
	}

	gfOutputFp, err := getOutputFilePath(runOutputDir)
	if err != nil {
		return nil, err
	}
	gfRecordFp, err := getRecordFilePath(runOutputDir)
	if err != nil {
		return nil, err
	}
	gfErrFp, err := getErrFilePath(runOutputDir)
	if err != nil {
		return nil, err
	}

	boolFirstRun := input.Note == NotePrintInput

	// Create the input file into disk
	SerializeInput(input, gfInputFp)

	var globalTuple string
	if GlobalTuple {
		globalTuple = "1"
	} else {
		globalTuple = "0"
	}

	var cmd *exec.Cmd
	if task.input.TestName != "" {
		cmd = exec.Command("go", "test", "-v", "-run", input.TestName, "./...")
	} else if task.input.CustomCmd != "" {
		cmds := strings.SplitN(task.input.CustomCmd, " ", 2)
		cmd = exec.Command(cmds[0], cmds[1])
	} else {
		return nil, fmt.Errorf("either testname or custom command is required")
	}
	cmd.Dir = TargetGoModDir

	// setting up environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("GF_RECORD_FILE=%s", gfRecordFp))
	env = append(env, fmt.Sprintf("GF_INPUT_FILE=%s", gfInputFp))
	env = append(env, fmt.Sprintf("BitGlobalTuple=%s", globalTuple))
	cmd.Env = env

	// setting up redirection
	var stdErrBuf, stdOutBuf bytes.Buffer
	cmd.Stdout = &stdOutBuf
	cmd.Stderr = &stdErrBuf
	runErr := cmd.Run()

	// Save output to the file
	outputF, err := os.Create(gfOutputFp)
	if err != nil {
		return nil, fmt.Errorf("create stdout: %s", err)
	}
	defer outputF.Close()
	outputW := bufio.NewWriter(outputF)
	outputW.Write(stdOutBuf.Bytes())
	outputW.Flush()

	// Save error to the file
	errF, err := os.Create(gfErrFp)
	if err != nil {
		return nil, fmt.Errorf("create stdout: %s", err)
	}
	defer errF.Close()
	errW := bufio.NewWriter(errF)
	errW.Write(stdErrBuf.Bytes())
	errW.Flush()

	if runErr != nil {
		return nil, fmt.Errorf("go test command fails: %s", err)
	}

	outputNumBug := CheckBugFromStdout(stdOutBuf.String())
	if outputNumBug != 0 {
		log.Println("Found a new bug. Now exit")
		os.Exit(1)
	}

	// Read the newly printed input file if this is the first run
	var retInput *Input
	if boolFirstRun {
		log.Printf("[Task %s] First run, reading input file %s", task.id, gfInputFp)
		bytes, err := ioutil.ReadFile(gfInputFp)
		if err != nil {
			return nil, err
		}

		retInput, err = ParseInputFile(string(bytes))
		if err != nil {
			return nil, err
		}

		// assign missing parts in input file
		retInput.TestName = task.input.TestName
		retInput.CustomCmd = task.input.CustomCmd

	} else {
		retInput = nil
	}

	// Read the printed record file
	b, err := ioutil.ReadFile(gfRecordFp)
	if err != nil {
		return nil, err
	}
	retRecord, err := ParseRecordFile(string(b))

	if err != nil {
		return nil, err
	}

	retOutput := NewRunResult()
	retOutput.RetInput = retInput
	retOutput.RetRecord = retRecord

	return retOutput, nil
}
