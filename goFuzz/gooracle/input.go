package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var MapInput map[string]runtime.SelectInfo
var SelectDelayMS int

func FileNameOfInput() string {
	return os.Getenv("GF_INPUT_FILE")
}

func ParseInputStr(text []string) map[string]runtime.SelectInfo {
	result := make(map[string]runtime.SelectInfo)

	strDelayMS := text[1]
	var err error
	SelectDelayMS, err = strconv.Atoi(strDelayMS)
	if err != nil {
		fmt.Println("The second line of input is not a number:", strDelayMS)
		return nil
	}

	for i, eachLine := range text {
		if i < 2 { // SelectDelayMS already stored
			continue
		}
		if eachLine == "" {
			continue
		}
		selectInput := runtime.SelectInfo{}
		// filename:linenum:totalCaseNum:chooseCaseNum
		vecStr := strings.Split(eachLine, ":")
		if len(vecStr) != 4 {
			fmt.Println("One line in input has incorrect format:", eachLine, "\tLine:", i)
			return nil
		}
		selectInput.StrFileName = vecStr[0]
		if _, err = strconv.Atoi(vecStr[1]); err != nil {
			fmt.Println("One line in input has incorrect format:", vecStr, "\tLine:", i)
			return nil
		}
		selectInput.StrLineNum = vecStr[1]
		if selectInput.IntNumCase, err = strconv.Atoi(vecStr[2]); err != nil {
			fmt.Println("One line in input has incorrect format:", vecStr, "\tLine:", i)
			return nil
		}

		if selectInput.IntPrioCase, err = strconv.Atoi(vecStr[3]); err != nil {
			fmt.Println("One line in input has incorrect format:", vecStr, "\tLine:", i)
			return nil
		}
		result[selectInput.StrFileName+":"+selectInput.StrLineNum] = selectInput
	}

	return result
}

func CreateInput() {
	out, err := os.Create(FileNameOfInput())
	if err != nil {
		fmt.Println("Failed to create file:", FileNameOfInput(), err)
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	var sb strings.Builder

	// The first line indicates that: if this input is used, then after the run a new input should be printed
	sb.WriteString(NotePrintInput)
	sb.WriteString("\n")

	// The second line is how many seconds to wait
	sb.WriteString(strconv.Itoa(SelectDelayMS))
	sb.WriteString("\n")

	// Each line corresponds to a select
	for _, selectInput := range runtime.MapSelectInfo {
		// filename:linenum:totalCaseNum:chooseCaseNum
		strFileName := selectInput.StrFileName
		if indexEnter := strings.Index(strFileName, "\n"); indexEnter > -1 {
			strFileName = strFileName[:indexEnter]
		}
		sb.WriteString(strFileName)
		sb.WriteString(":")
		sb.WriteString(selectInput.StrLineNum)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(selectInput.IntNumCase))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(selectInput.IntPrioCase))
		sb.WriteString("\n")
	}

	w.WriteString(sb.String())
}
