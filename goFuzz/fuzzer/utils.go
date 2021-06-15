package fuzzer

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// ListPackages lists all packages in the current module
// (Has to be run at the directory contains go.mod)
func ListPackages(goModRootPath string) ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	if goModRootPath != "" {
		cmd.Dir = goModRootPath

	}
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
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
func ListTestsInPackage(goModRootPath string, pkg string) ([]string, error) {
	if pkg == "" {
		pkg = "./..."
	}
	cmd := exec.Command("go", "test", "-list", ".*", pkg)
	if goModRootPath != "" {
		cmd.Dir = goModRootPath
	}
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	return parseGoCmdListOutput(out.String())
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
		if line != "" && strings.HasPrefix(line, "Test") {
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
