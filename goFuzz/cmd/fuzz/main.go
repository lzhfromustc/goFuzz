package main

import (
	"flag"
	"goFuzz/fuzzer"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	testBinsDir string
)

// parseFlag init logger and settings for the fuzzer
func parseFlag() {

	// Parse input
	pTargetGoModDir := flag.String("goModDir", "", "Directory contains Go Mod file")
	pTargetTestFunc := flag.String("testFunc", "", "Optional, if you only want to test single function in unit test")
	pTargetTestPkg := flag.String("testPkg", "", "Optional, fuzz tests under given dir")
	pOutputDir := flag.String("outputDir", "", "Full path of the output file")
	pModeGlobalTuple := flag.Bool("globalTuple", false, "Whether prev_location is global or per channel")
	maxParallel := flag.Int("parallel", 1, "Specified the maximum subroutine number for fuzzing.")
	scoreSdk := flag.Bool("scoreSdk", false, "Recording/scoring if channel comes from Go SDK")
	chCover := flag.String("chCover", "", "Optional. This parameter consumes a file path to a channel statistics file generated by printOperation before instrumentation.")
	pTargetTestBin := flag.String("testBin", "", "Use given binary file instead of calling go test")
	pBoolScoreAllPrim := flag.Bool("scoreAllPrim", false, "Recording/scoring other primitives like Mutex together with channel")
	pTimeDivide := flag.Int("timeDivideBy", 1, "Durations in time/sleep.go will be divided by this int number")
	pTestBinsDir := flag.String("testBinsDir", "", "")
	pSkipIntegration := flag.Bool("skipIntegration", false, "Should skip all the integration tests from the package.")
	flag.Parse()

	fuzzer.TargetTestFunc = *pTargetTestFunc
	fuzzer.OutputDir = *pOutputDir
	fuzzer.TargetGoModDir = *pTargetGoModDir
	fuzzer.GlobalTuple = *pModeGlobalTuple
	fuzzer.MaxParallel = *maxParallel
	fuzzer.ScoreSdk = *scoreSdk
	fuzzer.ScoreAllPrim = *pBoolScoreAllPrim
	fuzzer.OpCover = *chCover
	fuzzer.TargetTestPkg = *pTargetTestPkg
	fuzzer.TargetTestBin = *pTargetTestBin
	fuzzer.TimeDivide = *pTimeDivide
	fuzzer.SkipIntegration = *pSkipIntegration
	testBinsDir = *pTestBinsDir

	if fuzzer.OutputDir == "" {
		log.Fatal("-outputDir is required")
	}

	if _, err := os.Stat(fuzzer.OutputDir); os.IsNotExist(err) {
		err := os.Mkdir(fuzzer.OutputDir, os.ModePerm)
		if err != nil {
			log.Printf("create output folder failed: %v", err)
		}
	}

	if fuzzer.TargetGoModDir == "" && testBinsDir == "" {
		log.Fatal("Either -goModDir or -testBinsDir is required")
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
	var err error
	// parse command line flags
	parseFlag()

	goRoot, err := fuzzer.GetGoEnv("GOROOT")
	if err != nil {
		log.Fatalf("Failed to get go env GOROOT: %v", err)
	}
	fuzzer.GoRoot = goRoot
	log.Printf("current GOROOT: %s", goRoot)

	// setup logger
	setupLogger(filepath.Join(fuzzer.OutputDir, "fuzzer.log"))

	// print version
	log.Printf("Go Fuzzer Version: %s", fuzzer.Version)

	// find out which tests we need during this fuzzing
	var testsToFuzz []*fuzzer.GoTest
	if fuzzer.TargetTestFunc != "" {
		testsToFuzz = append(testsToFuzz, &fuzzer.GoTest{
			Func:    fuzzer.TargetTestFunc,
			Package: fuzzer.TargetTestPkg,
		})
	} else if testBinsDir != "" {
		tests, err := fuzzer.ListGoTestsFromFolderContainsTestBins(testBinsDir)
		if err != nil {
			log.Fatalf("failed to list tests by test bin at %s: %v", testBinsDir, err)
		}
		testsToFuzz = append(testsToFuzz, tests...)
	} else {
		log.Printf("finding all tests under module %s", fuzzer.TargetGoModDir)

		var packages []string
		if fuzzer.TargetTestPkg != "" {
			packages = append(packages, fuzzer.TargetTestPkg)
		} else {
			// Find all tests in all packages
			packages, err = fuzzer.ListPackages(fuzzer.TargetGoModDir)
			if err != nil {
				log.Fatalf("failed to list packages at %s: %v", fuzzer.TargetGoModDir, err)
			}

			log.Printf("found packages: %v", packages)
		}

		// ListTestsInPackage utilized command `go test -list` which cannot be run in parallel if they share same go code file.
		// Run parallel will cause `intput/output error` when `go test` tries to open file already opened by previous `go test` command.
		// Using other methold like `find Test | grep` can find test name but cannot find package location
		for _, pkg := range packages {

			testsInPkg, err := fuzzer.ListTestsInPackage(fuzzer.TargetGoModDir, pkg)
			if err != nil {
				log.Printf("[ignored] failed to list tests at package %s: %v", pkg, err)
				continue
			}

			for _, t := range testsInPkg {
				log.Printf("found %+v", *t)
			}

			testsToFuzz = append(testsToFuzz, testsInPkg...)

		}

	}

	// Parse operation statistics if need
	if fuzzer.OpCover != "" {
		numOfOpID, err := fuzzer.InitOperationStats(fuzzer.OpCover)
		if err != nil {
			log.Printf("[ignored] initialial channel coverage failed: %v", err)
			// continue fuzzing (ignore error)
			fuzzer.OpCover = ""
		} else {
			log.Printf("found %d operation IDs in %s", numOfOpID, fuzzer.OpCover)
		}
	}

	// Setup metrics streaming
	fuzzer.StreamMetrics(filepath.Join(fuzzer.OutputDir, "fuzzer-metrics.json"), 5)

	// Main entry for fuzzing
	fuzzer.Fuzz(testsToFuzz, nil, fuzzer.MaxParallel)
}
