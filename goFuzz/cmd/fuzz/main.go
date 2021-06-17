package main

import (
	"flag"
	"goFuzz/fuzzer"
	"io"
	"log"
	"os"
)

// parseFlag init logger and settings for the fuzzer
func parseFlag() {
	file, err := os.OpenFile("fuzzer.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	w := io.MultiWriter(file, os.Stdout)
	log.SetOutput(w)

	// Parse input
	pTargetGoModDir := flag.String("goModDir", "", "Directory contains Go Mod file")
	pTargetTestFunc := flag.String("testFunc", "", "Optional, if you only want to test single function in unit test")
	pOutputDir := flag.String("outputDir", "", "Full path of the output file")
	pModeGlobalTuple := flag.Bool("globalTuple", false, "Whether prev_location is global or per channel")
	maxParallel := flag.Int("parallel", 1, "Specified the maximum subroutine number for fuzzing.")

	flag.Parse()

	fuzzer.TargetTestFunc = *pTargetTestFunc
	fuzzer.OutputDir = *pOutputDir
	fuzzer.TargetGoModDir = *pTargetGoModDir
	fuzzer.GlobalTuple = *pModeGlobalTuple
	fuzzer.MaxParallel = *maxParallel

	if fuzzer.OutputDir == "" {
		log.Fatal("-outputDir is required")
	}

	if fuzzer.TargetGoModDir == "" {
		log.Fatal("-goModDir is required")
	}
}

func main() {
	parseFlag()

	var tests []string
	var err error
	if fuzzer.TargetTestFunc != "" {
		tests = append(tests, fuzzer.TargetTestFunc)
	} else {
		tests, err = fuzzer.ListTestsInPackage(fuzzer.TargetGoModDir, "./...")
		if err != nil {
			log.Fatal(err)
		}
	}

	// Main entry for fuzzing
	fuzzer.Fuzz(tests, nil, fuzzer.MaxParallel)
}
