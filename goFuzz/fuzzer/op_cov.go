package fuzzer

import (
	"io/ioutil"
	"log"
	"strings"
)

var (
	// opID2Type is the map from operation ID to operation type (chmake, chsend, etc..)
	opID2Type   map[string]string
	totalReport OpCovReport
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
	ids := []string{}
	for id, _ := range opID2Type {
		ids = append(ids, id)
	}

	totalReport = GetOperationCoverageReport(opID2Type, ids)

	// since 0 cannot be divided
	if totalReport.numOfChOp == 0 {
		totalReport.numOfChOp = 1
	}

	if totalReport.numOfChMake == 0 {
		totalReport.numOfChOp = 1
	}

	if totalReport.numOfOtherPrimitivesOp == 0 {
		totalReport.numOfChOp = 1
	}

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

type OpCovReport struct {
	numOfChMake            uint
	numOfChOp              uint
	numOfOtherPrimitivesOp uint
}

func GetOperationCoverageReport(totalID2Type map[string]string, ids []string) OpCovReport {
	totalNumOfCh := len(totalID2Type)
	report := OpCovReport{}

	if totalNumOfCh == 0 {
		return report
	}

	numOfMatchedID := 0
	recorded := make(map[string]bool)

	for _, id := range ids {
		if _, dup := recorded[id]; dup {
			continue
		}
		recorded[id] = true

		t, exist := totalID2Type[id]
		if exist {
			switch t {
			case "chmake":
				report.numOfChMake += 1
			case "chsend", "chrecv", "chclose":
				report.numOfChOp += 1
			default:
				report.numOfOtherPrimitivesOp += 1
			}
			numOfMatchedID += 1
		}
	}

	return report
}

func PrintOperationCoverageReport(totalReport OpCovReport, currReport OpCovReport) {
	log.Printf("channel make count %d, coverage %.2f%%", currReport.numOfChMake, float32(currReport.numOfChMake)/float32(totalReport.numOfChMake))
	log.Printf("channel op count %d,coverage %.2f%%", currReport.numOfChOp, float32(currReport.numOfChOp)/float32(totalReport.numOfChOp))
	log.Printf("other primitive op count %d,coverage %.2f%%", currReport.numOfOtherPrimitivesOp, float32(currReport.numOfOtherPrimitivesOp)/float32(totalReport.numOfOtherPrimitivesOp))
}
