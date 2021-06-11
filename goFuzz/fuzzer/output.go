package fuzzer

import (
	"strings"
)

type RunOutput struct {
	RetInput  *Input
	RetRecord *Record
	// TODO:: Is there a better data structure than using string directly? enum?
	Stage string // "unknown", "deter", "calib" or "rand"
}

func ParseOutputFile(content string) (numBug int) {

	text := strings.Split(content, "\n")

	for _, oneLine := range text {
		if oneLine == "-----New Bug:" {
			numBug++
		}
	}

	return
}
