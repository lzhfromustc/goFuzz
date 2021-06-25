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

			targetLine = strings.TrimLeft(targetLine, " \t")
			parts := strings.Split(targetLine, " ")

			if len(parts) != 2 {
				return nil, fmt.Errorf("malformed stacktrace, expected format: <file>:<line> <stack offset>, got %s", targetLine)
			}
			ids = append(ids, parts[0])
		}
	}

	return ids, nil
}
