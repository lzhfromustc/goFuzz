package config

import "time"

var FuzzerDeadline time.Duration = time.Minute * 15
var FuzzerSelectDelayVector []int = []int{500, 2000, 8000}
