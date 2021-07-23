package fuzzer

import (
	"fmt"
	"strings"
)

// GetListOfBugIDFromStdoutContent parses gooracle related information(Bug ID) from stdout file content.
// A Bug ID is where channel was been created (identified by file path + line number) and
// that channel will trigger block bug.
//
// Example from stdout
// -----New Blocking Bug:
// goroutine 3855 [running]:
// github.com/prometheus/prometheus/tsdb/wal.(*WAL).run(0xc0002e7c20)
// /Users/xsh/code/prometheus/tsdb/wal/wal.go:372 +0x47a <------ "/Users/xsh/code/prometheus/tsdb/wal/wal.go:372" is Bug ID
//
func GetListOfBugIDFromStdoutContent(c string) ([]string, error) {
	lines := strings.Split(c, "\n")
	ids := make([]string, 0)
	numOfLines := len(lines)
	for idx, line := range lines {
		if line == "" {
			continue
		}

		// trim space and tab
		line = strings.TrimLeft(line, " \t")
		if strings.HasPrefix(line, "-----New Blocking Bug:") || strings.HasPrefix(line, "-----New NonBlocking Bug:") { // Note for Shihao: changed the format of bug report; now reporting the type of bug

			// skip `goroutine 3855 [running]:` and package based function
			// to get line contains filesystem location
			idLineIdx := idx + 3

			// Skip file location(s) that is belongs to my*.go until find the bug root cause
			for {
				if idLineIdx >= numOfLines {
					return nil, fmt.Errorf("total line %d, target bug ID line at %d", numOfLines, idLineIdx)
				}

				if strings.Contains(lines[idLineIdx], "src/runtime/my") {
					idLineIdx += 2
				}

				// if this line is not from our my*.go files, then it is where bug happened
				break
			}

			targetLine := lines[idLineIdx]

			id, err := getFileAndLineFromStacktraceLine(targetLine)

			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}

		if strings.HasPrefix(line, "panic") {
			// panic: send on closed channel
			//
			// goroutine 7 [running]:
			// fuzzer-toy/blocking/grpc/1353.(*roundRobin).watchAddrUpdates(0xc00001c810)
			//     /fuzz/target/blocking/grpc/1353/grpc1353_test.go:84 +0x10f <==== idLineIdx
			// ....
			idLineIdx := idx + 4

			// Skip file location(s) that is belongs to my*.go until find the bug root cause
			for {
				if idLineIdx >= numOfLines {
					return nil, fmt.Errorf("total line %d, target bug ID line at %d", numOfLines, idLineIdx)
				}

				if strings.Contains(lines[idLineIdx], "src/runtime/my") {
					idLineIdx += 2
				}

				// if this line is not from our my*.go files, then it is where bug happened
				break
			}

			targetLine := lines[idLineIdx]

			id, err := getFileAndLineFromStacktraceLine(targetLine)

			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// getFileAndLineFromStacktraceLine returns only <file>:<line>
// from string with format <file>:<line> <stack offset>
func getFileAndLineFromStacktraceLine(line string) (string, error) {
	targetLine := strings.TrimLeft(line, " \t")
	parts := strings.Split(targetLine, " ")

	if len(parts) != 2 {
		return "", fmt.Errorf("malformed stacktrace, expected format: <file>:<line> <stack offset>, got %s", targetLine)
	}

	return parts[0], nil
}
