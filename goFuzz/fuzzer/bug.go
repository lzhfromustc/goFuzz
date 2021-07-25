package fuzzer

import (
	"fmt"
	"log"
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
// 	/Users/xsh/code/prometheus/tsdb/wal/wal.go:372 +0x47a <------ "/Users/xsh/code/prometheus/tsdb/wal/wal.go:372" is Bug ID
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
		if strings.HasPrefix(line, "-----New Blocking Bug:") || strings.HasPrefix(line, "-----New NonBlocking Bug:") ||
			strings.HasPrefix(line, "panic") || strings.HasPrefix(line, "fatal error") {

			if strings.HasPrefix(line, "panic: test timed out after") {
				continue
			}
			idLineIdx := idx + 1

			// Skip file location(s) that is belongs to my*.go until find the bug root cause
			// if this line is not from our my*.go files, then it is where bug happened
			for {
				if idLineIdx >= numOfLines {
					return nil, fmt.Errorf("total line %d, target bug ID line at %d", numOfLines, idLineIdx)
				}

				if strings.Contains(lines[idLineIdx], "src/runtime/my") || strings.Contains(lines[idLineIdx], "src/sync") {
					idLineIdx += 2
					continue
				}

				if strings.HasPrefix(lines[idLineIdx], "\t") {
					// first line with filename + linenumer
					break
				}

				idLineIdx += 1
			}

			targetLine := lines[idLineIdx]

			id, err := getFileAndLineFromStacktraceLine(targetLine)

			if err != nil {
				log.Printf("getFileAndLineFromStacktraceLine failed: %s", err)
				continue
			}
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// getFileAndLineFromStacktraceLine returns only <file>:<line>
// from string with format <file>:<line> [<stack offset>]
func getFileAndLineFromStacktraceLine(line string) (string, error) {
	targetLine := strings.TrimLeft(line, " \t")
	parts := strings.Split(targetLine, " ")
	fileAndLine := parts[0]

	if len(strings.Split(fileAndLine, ":")) != 2 {
		return "", fmt.Errorf("malformed stacktrace, expected format: <file>:<line> [<stack offset>], got %s", targetLine)
	}

	return fileAndLine, nil
}
