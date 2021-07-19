package fuzzer

import (
	"io/ioutil"
	"log"
	"strings"
)

type OpCovReport struct {
	numOfChMake            uint
	numOfChOp              uint
	numOfOtherPrimitivesOp uint
}

var (
	// opID2Type is the map from operation ID to operation type (chmake, chsend, etc..)
	opID2Type     map[string]string
	totalReport   OpCovReport
	triggeredOpID map[string]bool
)

// InitOperationStats open and parse the file contains operation statistics
// It returns the number of operation ID
func InitOperationStats(chStatFile string) (int, error) {
	triggeredOpID = make(map[string]bool)
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
	for id := range opID2Type {
		ids = append(ids, id)
	}

	totalReport = GetCurrOpIDCoverageReport(opID2Type, ids)

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

func UpdateTriggeredOpID(triggeredIDs map[string]bool, ids []string) {
	if triggeredIDs == nil {
		return
	}
	for _, id := range ids {
		if _, exist := triggeredIDs[id]; !exist {
			triggeredIDs[id] = true
		}
	}
}

func GetTriggeredOpIDCoverageReport(totalID2Type map[string]string, triggeredIDinTotal map[string]bool) OpCovReport {
	totalNumOfCh := len(totalID2Type)
	report := OpCovReport{}

	if totalNumOfCh == 0 || triggeredIDinTotal == nil {
		return report
	}

	for id := range triggeredIDinTotal {

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
		}
	}

	return report
}

func GetCurrOpIDCoverageReport(totalID2Type map[string]string, ids []string) OpCovReport {
	totalNumOfCh := len(totalID2Type)
	report := OpCovReport{}

	if totalNumOfCh == 0 {
		return report
	}

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
		}
	}

	return report
}

func PrintCurrOpIDCovReport(totalReport OpCovReport, currReport OpCovReport) {
	log.Printf("channel make count %d, coverage %.2f%%", currReport.numOfChMake, float32(currReport.numOfChMake)/float32(totalReport.numOfChMake)*100)
	log.Printf("channel op count %d,coverage %.2f%%", currReport.numOfChOp, float32(currReport.numOfChOp)/float32(totalReport.numOfChOp)*100)
	log.Printf("other primitive op count %d,coverage %.2f%%", currReport.numOfOtherPrimitivesOp, float32(currReport.numOfOtherPrimitivesOp)/float32(totalReport.numOfOtherPrimitivesOp)*100)
}

func PrintTriggeredOpIDCovReport(totalReport OpCovReport, triggeredReport OpCovReport) {
	log.Printf("cumulative channel make count %d, coverage %.2f%%", triggeredReport.numOfChMake, float32(triggeredReport.numOfChMake)/float32(totalReport.numOfChMake)*100)
	log.Printf("cumulative channel op count %d,coverage %.2f%%", triggeredReport.numOfChOp, float32(triggeredReport.numOfChOp)/float32(totalReport.numOfChOp)*100)
	log.Printf("cumulative other primitive op count %d,coverage %.2f%%", triggeredReport.numOfOtherPrimitivesOp, float32(triggeredReport.numOfOtherPrimitivesOp)/float32(totalReport.numOfOtherPrimitivesOp)*100)
}
