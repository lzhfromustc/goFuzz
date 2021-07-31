"""
This file is to calculate the output of command like `go test ./... -coverprofile=coverage.out`
in case of go tool cover failed to parse `coverage.out` file
"""
import sys
import re
def main():
    file = sys.argv[1]
    cov_key = "coverage: "
    total = 0
    cnt = 0
    with open(file) as f:
        lines = list(f)
        lines = [line for line in lines if line.find("coverage: ") != -1]
        for line in lines:
            cov_idx = line.find(cov_key)
            num_str = line[cov_idx+len(cov_key):line.find("%")]
            cnt += 1
            if num_str[0] == '[':
                continue
        
            total += float(num_str)
    print(f"{total/cnt:.2f}%")

if __name__ == "__main__":
    main()