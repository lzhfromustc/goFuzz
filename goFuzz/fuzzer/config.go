package fuzzer

import "time"

var (
	FuzzerDeadline          time.Duration = time.Minute * 15
	FuzzerSelectDelayVector []int         = []int{500, 2000, 8000}
	TargetGoModDir          string        // Directory contains Go Mod file
	TargetTestFunc          string        // Optional, if you only want to run single test instead of whole tests, regex also accept (will be the argument of go test -run)
	OutputDir               string        // Output directory, each folder contains output(stdout), record (generated by gooracle) and input (generated by fuzzer/changed by goorcle at first run)
	GlobalTuple             bool          // Recording executed select combination in global or not (per Goroutine)
	MaxParallel             int           // Max Parallel worker (how many fuzz target can be run at the same time)
	ScoreSdk                bool          // recording/scoring if channel comes from Go SDK
	OpCover                 string        // File path of channel statistics (generated by printOperation)
	TargetTestDir           string
)
