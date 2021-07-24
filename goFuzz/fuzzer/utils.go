package fuzzer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ListPackages lists all packages in the current module
// (Has to be run at the directory contains go.mod)
func ListPackages(goModRootPath string) ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	if goModRootPath != "" {
		cmd.Dir = goModRootPath
	}
	cmd.Env = os.Environ()

	var out bytes.Buffer
	w := io.MultiWriter(&out, log.Writer())
	cmd.Stdout = w
	cmd.Stderr = w

	log.Printf("go list ./... in %s", goModRootPath)
	err := cmd.Run()

	if err != nil {
		return nil, fmt.Errorf("[go list ./...] %v", err)
	}

	return parseGoCmdListOutput(out.String())

}

func parseGoCmdListOutput(output string) ([]string, error) {
	lines := strings.Split(output, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}

// ListTestsInPackage lists all tests in the given package
// pkg can be ./... to search in all packages
func ListTestsInPackage(goModRootPath string, pkg string) ([]*GoTest, error) {
	if pkg == "" {
		pkg = "./..."
	}

	// prepare timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "-list", ".*", pkg)
	if goModRootPath != "" {
		cmd.Dir = goModRootPath
	}
	cmd.Env = os.Environ()

	var out bytes.Buffer
	w := io.MultiWriter(&out, log.Writer())
	cmd.Stdout = w
	cmd.Stderr = w

	log.Printf("go test -list .* %s", pkg)

	err := cmd.Run()

	if err != nil {
		return nil, fmt.Errorf("[go test -list .* %s] %v", pkg, err)
	}

	testFuncs, err := parseGoCmdTestListOutput(out.String())
	if err != nil {
		return nil, err
	}

	goTests := make([]*GoTest, 0, len(testFuncs))
	for _, testFunc := range testFuncs {
		goTests = append(goTests, &GoTest{
			Func:    testFunc,
			Package: pkg,
		})
	}
	return goTests, nil
}

func parseGoCmdTestListOutput(output string) ([]string, error) {
	lines := strings.Split(output, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		// To filter out output likes
		// ?   	k8s.io/kubernetes/cluster/images/etcd-version-monitor	[no test files]
		// ok      goFuzz/example/simple1  0.218s
		// Only keep output like:
		// TestParseInputFileHappy

		if line != "" && strings.HasPrefix(line, "Test") && line != "Test" && !strings.ContainsAny(line, " ") {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// createDir create an folder (create folder if not exist)
func createDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	} else {
		// return any other error if occurs
		return err
	}
}

func GetGoEnv(key string) (string, error) {
	cmd := exec.Command("go", "env", key)
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return strings.TrimRight(out.String(), "\n"), nil
}

const float64EqualityThreshold = 1e-9

func equal64(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}
