#!/usr/bin/env python3
import argparse
import glob
import subprocess
from typing import List
import pathlib
import os

PROJ_ROOT_DIR = pathlib.Path(__file__).parent.parent.absolute().as_posix()

def find_gotest_in_folder(folder: str) -> List[str]:
    """Find all *_test.go files under the given folder

    Args:
        folder (str): A foder path

    Returns:
        [str]: A list of file path
    """
    if folder.endswith("/"):
        folder = folder[:-1]
    glob_path = os.path.join(folder, '**', '*.go')
    print(f"using glob {glob_path}")
    return glob.glob(glob_path, recursive=True)

def instrument_gotest(file: str):
    """Instrument Golang test file by bin/instrument

    Args:
        file (str): Golang test file going to be instrumented
    """
    inst_file = os.path.join(PROJ_ROOT_DIR, "bin/instrument")
    if not os.path.exists(inst_file):
        raise Exception("Please run 'make instrument' to generate target binary file.")
    res = subprocess.run([inst_file, f"-file={file}"])
    res.check_returncode()
    print(f"Instrumented file {file}")

def main():
    parser = argparse.ArgumentParser(description='Instrument Golang test files.')
    parser.add_argument("dir", metavar='DIR', type=str, nargs='+',
                        help='directory contains *_test.go files')

    args = parser.parse_args()
    for dir in args.dir:
        files = find_gotest_in_folder(dir)

        for f in files:
            abs_f = os.path.abspath(f)
            instrument_gotest(abs_f)



if __name__ == "__main__":
    main()