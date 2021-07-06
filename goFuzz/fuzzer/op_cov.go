package fuzzer

import (
	"io/ioutil"
	"strings"
)

// InitChStats open and parse the file contains operation statistics
// It returns the number of operation ID
func InitOperationStats(chStatFile string) (int, error) {
	bytes, err := ioutil.ReadFile(chStatFile)
	if err != nil {
		return 0, err
	}
	ids, err := parseOperationCoverageFileContent(string(bytes))
	if err != nil {
		return 0, nil
	}

	fuzzerContext.opStats = ids

	return len(ids), nil
}

// parseOperationCoverageFileContent parses the operation statistics file content and
// returns a list of channel ID
func parseOperationCoverageFileContent(content string) ([]string, error) {
	lines := strings.Split(content, "\n")
	var chIDs []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		chIDs = append(chIDs, line)
	}

	return chIDs, nil
}

// GetOperationCoverage calculates the percentage of operations(`pmIDs`) in the `chStats`
func GetOperationCoverage(totalPmIDs []string, pmIDs []string) float32 {
	totalNumOfCh := len(totalPmIDs)

	if totalNumOfCh == 0 {
		return 0
	}

	numOfMatchedCh := 0

	for _, id := range pmIDs {
		if contains(totalPmIDs, id) {
			numOfMatchedCh += 1
		}
	}

	return float32(numOfMatchedCh) / float32(totalNumOfCh)
}
