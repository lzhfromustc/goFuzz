package fuzzer

import (
	"io/ioutil"
	"log"
	"strings"
)

var (
	// opID2Type is the map from operation ID to operation type (chmake, chsend, etc..)
	opID2Type map[string]string
)

// InitOperationStats open and parse the file contains operation statistics
// It returns the number of operation ID
func InitOperationStats(chStatFile string) (int, error) {
	bytes, err := ioutil.ReadFile(chStatFile)
	if err != nil {
		return 0, err
	}
	res, err := parseOperationCoverageFileContent(string(bytes))
	if err != nil {
		return 0, nil
	}

	opID2Type = res

	return len(res), nil
}

// parseOperationCoverageFileContent parses the operation statistics file content and
// returns a list of channel ID
func parseOperationCoverageFileContent(content string) (map[string]string, error) {
	lines := strings.Split(content, "\n")
	var id2type = make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			log.Printf("[ignored] malformed format of operation cov: %s", line)
			continue
		}
		id2type[parts[0]] = parts[1]
	}

	return id2type, nil
}

// GetOperationCoverage calculates the percentage of operations(`pmIDs`) in the `chStats`
func GetOperationCoverage(totalID2Type map[string]string, ids []string) float32 {
	totalNumOfCh := len(totalID2Type)

	if totalNumOfCh == 0 {
		return 0
	}

	numOfMatchedID := 0

	for _, id := range ids {

		_, exist := totalID2Type[id]
		if exist {
			numOfMatchedID += 1
		}
	}

	return float32(numOfMatchedID) / float32(totalNumOfCh)
}
