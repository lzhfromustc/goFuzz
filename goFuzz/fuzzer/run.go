package fuzzer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
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

func getOpCovFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "opcov"))
}

func getRecordFilePath(outputDir string) (string, error) {
	return filepath.Abs(path.Join(outputDir, "record"))
}

func NewRunTask(input *Input, stage FuzzStage, entryIdx uint64, execCount int, entry *FuzzQueryEntry) (*RunTask, error) {
	var mainName string

	if input.GoTestCmd != nil {
		if input.GoTestCmd.Package != "" {
			// Whole package URL might be too long for filesystem, so only use last component of the package URL.
			basePkg := path.Base(input.GoTestCmd.Package)
			mainName = fmt.Sprintf("%s-%s", basePkg, input.GoTestCmd.Func)
		} else {
			mainName = input.GoTestCmd.Func
		}
	} else if input.CustomCmd != "" {
		mainName = input.CustomCmd
	} else {
		return nil, errors.New("malformed input, either GoTestCmd or CustomCmd is required")
	}

	task := &RunTask{
		input: input,
		entry: entry,
		stage: stage,
		id:    fmt.Sprintf("%s-%s-%d-%d", mainName, stage, entryIdx, execCount),
	}
	return task, nil
}

func Run(fuzzCtx *FuzzContext, task *RunTask) (*RunResult, error) {
	var err error
	input := task.input

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
	gfOpCovFp, err := getOpCovFilePath(runOutputDir)
	if err != nil {
		return nil, err
	}

	boolFirstRun := input.Note == NotePrintInput

	// Create the input file into disk
	err = SerializeInput(input, gfInputFp)
	if err != nil {
		return nil, err
	}

	var globalTuple string
	if GlobalTuple {
		globalTuple = "1"
	} else {
		globalTuple = "0"
	}

	// prepare timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if task.input.GoTestCmd != nil {
		if TargetTestBin != "" {
			// Since golang's compiled test can only be one per package, so we just assume the test func must exist in the given binary
			cmd = exec.CommandContext(ctx, TargetTestBin, "test.parallel", "1", "-test.v", "-test.run", input.GoTestCmd.Func)
		} else {
			var pkg = input.GoTestCmd.Package
			if pkg == "" {
				pkg = "./..."
			}
			cmd = exec.CommandContext(ctx, "go", "test", "-v", "-run", input.GoTestCmd.Func, pkg)
		}
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
	env = append(env, fmt.Sprintf("GF_OP_COV_FILE=%s", gfOpCovFp))
	env = append(env, fmt.Sprintf("BitGlobalTuple=%s", globalTuple))
	if ScoreSdk {
		env = append(env, "GF_SCORE_SDK=1")
	}
	if ScoreAllPrim {
		env = append(env, "GF_SCORE_TRAD=1")
	}
	if GoRoot != "" {
		env = append(env, fmt.Sprintf("GOROOT=%s", GoRoot))
	}

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
		// Go test failed might be intentional
		log.Printf("[Task %s][ignored] go test command failed: %v", task.id, err)
	}

	// Read the newly printed input file if this is the first run
	var retInput *Input
	if boolFirstRun {
		log.Printf("[Task %s] first run, reading input file %s", task.id, gfInputFp)
		bytes, err := ioutil.ReadFile(gfInputFp)
		if err != nil {
			return nil, err
		}

		retInput, err = ParseInputFile(string(bytes))
		if err != nil {
			return nil, err
		}

		// assign missing parts in input file
		retInput.GoTestCmd = task.input.GoTestCmd
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

	// Read bug IDs from stdout
	bugIDs, err := GetListOfBugIDFromStdoutContent(stdOutBuf.String())
	if err != nil {
		return nil, err
	}

	// Read executed operations from gfOpCovFp
	b, err = ioutil.ReadFile(gfOpCovFp)
	if err != nil {
		// if error happened, log and ignore
		log.Printf("[Task %s][ignored] cannot read operation coverage file %s: %v", task.id, gfOpCovFp, err)
	}
	ids := strings.Split(string(b), "\n")

	retOutput := &RunResult{
		RetInput:       retInput,
		RetRecord:      retRecord,
		BugIDs:         bugIDs,
		StdoutFilepath: gfOutputFp,
		opIDs:          ids,
	}
	retOutput.RetInput = retInput
	retOutput.RetRecord = retRecord

	// Increment number of runs after a successfully run
	fuzzCtx.IncNumOfRun()
	return retOutput, nil
}
