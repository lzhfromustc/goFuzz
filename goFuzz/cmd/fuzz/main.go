package main

import (
	"flag"
	"goFuzz/fuzzer"
	"io"
	"log"
	"os"
	"path/filepath"
)

// parseFlag init logger and settings for the fuzzer
func parseFlag() {

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

	if _, err := os.Stat(fuzzer.OutputDir); os.IsNotExist(err) {
		err := os.Mkdir(fuzzer.OutputDir, os.ModePerm)
		if err != nil {
			log.Printf("create output folder failed: %v", err)
		}
	}

	if fuzzer.TargetGoModDir == "" {
		log.Fatal("-goModDir is required")
	}
}

func setupLogger(logFile string) {

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	w := io.MultiWriter(file, os.Stdout)
	log.SetOutput(w)
}

func main() {

	// parse command line flags
	parseFlag()

	// setup logger
	setupLogger(filepath.Join(fuzzer.OutputDir, "fuzzer.log"))

	// find out which tests we need during this fuzzing
	var testsToFuzz []*fuzzer.GoTest
	if fuzzer.TargetTestFunc != "" {
		testsToFuzz = append(testsToFuzz, &fuzzer.GoTest{
			Func: fuzzer.TargetTestFunc,
		})
	} else {
		log.Printf("finding all tests under module %s", fuzzer.TargetGoModDir)
		// Find all tests in all packages
		packages, err := fuzzer.ListPackages(fuzzer.TargetGoModDir)
		if err != nil {
			log.Fatalf("failed to list package at %s: %v", fuzzer.TargetGoModDir, err)
		}

		for _, pkg := range packages {
			testsInPkg, err := fuzzer.ListTestsInPackage(fuzzer.TargetGoModDir, pkg)
			if err != nil {
				log.Fatalf("failed to list tests at package %s: %v", pkg, err)
			}

			for _, t := range testsInPkg {
				log.Printf("found %+v", *t)
			}

			testsToFuzz = append(testsToFuzz, testsInPkg...)
		}
	}

	// Setup metrics streaming
	fuzzer.StreamMetrics(filepath.Join(fuzzer.OutputDir, "fuzzer-metrics.json"), 5)

	// Main entry for fuzzing
	fuzzer.Fuzz(testsToFuzz, nil, fuzzer.MaxParallel)
}
