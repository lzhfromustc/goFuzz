package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Parse input
	pProjectPath := flag.String("path", "", "Full path of the target project")
	pProjectGOPATH := flag.String("GOPATH", "", "GOPATH of the target project")
	pOutputFullPath := flag.String("output", "", "Full path of the output file")

	flag.Parse()

	strTestPath := *pProjectPath
	strProjectGOPATH := *pProjectGOPATH
	strOutputFullPath := *pOutputFullPath

	err := os.Setenv("GOPATH", strProjectGOPATH)
	if err != nil {
		fmt.Println("The export of GOPATH fails:", err)
		return
	}
	err = os.Setenv("OutputFullPath", strOutputFullPath)
	if err != nil {
		fmt.Println("The export of OutputFullPath fails:", err)
		return
	}
	err = os.Setenv("TestMod", "TestOnce")
	if err != nil {
		fmt.Println("The export of TestMod=TestOnce fails:", err)
		return
	}

	cmd := exec.Command("go", "test", strTestPath+"/...")
	err = cmd.Run()
		if err != nil {
		fmt.Println("The go test command fails:", err)
		return
	}


}