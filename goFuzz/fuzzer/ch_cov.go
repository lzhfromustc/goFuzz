package fuzzer

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// InitChStats open and parse the file contains channel statistics
// It returns the a list of channel ID (format is filename:line)
func InitChStats(chStatFile string) error {
	bytes, err := ioutil.ReadFile(chStatFile)
	if err != nil {
		return err
	}
	chIDs, err := parseChStatsFileContent(string(bytes))
	if err != nil {
		return nil
	}

	fuzzerContext.chStats = chIDs

	return nil
}

// parseChStatsFileContent parses the channel statistics file content and
// returns a list of channel ID
func parseChStatsFileContent(content string) ([]string, error) {
	lines := strings.Split(content, "\n")
	var chIDs []string
	for _, line := range lines {
		if line == "" {
			continue
		}

		// check basic formats
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected format in channel statistics is filename:line, actual: %s", line)
		}
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("expected format in channel statistics is filename:line, actual: %s", line)
		}
		if i < 0 {
			return nil, fmt.Errorf("expected format in channel statistics is filename:line, actual: %s", line)
		}
		chIDs = append(chIDs, line)
	}

	return chIDs, nil
}

// GetChannelCoverage calculates the percentage of channels in the `records` in the `chStats`
func GetChannelCoverage(chStats []string, records []ChanRecord) float32 {
	totalNumOfCh := len(chStats)

	if totalNumOfCh == 0 {
		return 0
	}

	numOfMatchedCh := 0

	for _, ch := range records {
		if contains(chStats, ch.ChID) {
			numOfMatchedCh += 1
		}
	}

	return float32(numOfMatchedCh) / float32(totalNumOfCh)
}
