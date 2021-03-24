package fuzzer

import (
	"bufio"
	"goFuzz/config"
	"os"
)

func ParseOutputFile() (numBug int) {

	file, err := os.Open(config.StrOutputFullPath)
	if err != nil { // This may be the first run
		//fmt.Println("Failed to open output file:", config.StrOutputFullPath)
		return 0
	}
	defer file.Close()

	var text []string

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	for _, oneLine := range text {
		if oneLine == "-----New Bug:" {
			numBug++
		}
	}

	return
}