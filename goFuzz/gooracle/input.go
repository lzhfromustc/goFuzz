package gooracle

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func FileNameOfInput() string {
	return StrTestpath + "/" + InputFileName
}

func ParseInputStr(text []string) map[string]runtime.SelectInput {
	result := make(map[string]runtime.SelectInput)

	strDelayMS := text[1]
	var err error
	runtime.SelectDelayMS, err = strconv.Atoi(strDelayMS)
	if err != nil {
		fmt.Println("The first line of input is not a number:", strDelayMS)
		return nil
	}

	for i, eachLine := range text {
		if i < 2 { // SelectDelayMS already stored
			continue
		}
		if eachLine == "" {
			continue
		}
		selectInput := runtime.SelectInput{}
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
		result[selectInput.StrFileName + ":" + selectInput.StrLineNum] = selectInput
	}

	return result
}

func CreateInput() {
	out, err := os.Create(FileNameOfInput())
	if err != nil {
		fmt.Println("Failed to create file:", FileNameOfInput())
		return
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	str := ""
	// The first line indicates that: if this input is used, then after the run a new input should be printed
	str += NotePrintInput + "\n"

	// The second line is how many seconds to wait
	str += strconv.Itoa(runtime.SelectDelayMS) + "\n"

	// Each line corresponds to a select
	for _, selectInput := range runtime.MapFirstInput {
		// filename:linenum:totalCaseNum:chooseCaseNum
		strFileName := selectInput.StrFileName
		if indexEnter := strings.Index(strFileName, "\n"); indexEnter > -1 {
			strFileName = strFileName[:indexEnter]
		}
		str += strFileName + ":" + selectInput.StrLineNum
		str += ":" + strconv.Itoa(selectInput.IntNumCase)
		str += ":" + strconv.Itoa(selectInput.IntPrioCase)
		str += "\n"
	}

	w.WriteString(str)
}
