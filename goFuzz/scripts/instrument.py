#!/usr/bin/env python3
import argparse
import glob
import subprocess
from typing import List
import pathlib
import os

PROJ_ROOT_DIR = pathlib.Path(os.path.realpath(__file__)).parent.parent.as_posix()
BIN_INSTRUMENT = None
BIN_PRINTOPERATION = None


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

def generate_ch_statistics(bin:str, go_file:str, output_file:str):
    """Dump channel statistics to the given output file by bin/printOperation
    :param file:
    :param output:
    :return:
    """
    res = subprocess.run([bin, f"-file={go_file}", f"-output={output_file}"])
    res.check_returncode()
    print(f"Dump channel statistics of file {go_file} to {output_file}")

def instrument_gotest(bin:str, go_file: str, rec_out_file:str):
    """Instrument Golang test file by bin/instrument

    Args:
        file (str): Golang test file going to be instrumented
    """
    res = subprocess.run([bin, f"-file={go_file}", f"-output={rec_out_file}"])
    res.check_returncode()
    print(f"Instrumented file {go_file}")

def main():
    parser = argparse.ArgumentParser(description='Instrument Golang test files.')
    parser.add_argument("dir", metavar='DIR', type=str, nargs='+',
                        help='directory contains *_test.go files')
    parser.add_argument("--op-out", help="Output path of channel statistics(of Go source code in the given dir) ")

    args = parser.parse_args()
    op_out = args.op_out

    global PROJ_ROOT_DIR
    global BIN_INSTRUMENT
    global BIN_PRINTOPERATION

    BIN_INSTRUMENT = os.path.join(PROJ_ROOT_DIR, "bin/instrument")
    BIN_PRINTOPERATION = os.path.join(PROJ_ROOT_DIR, "bin/printOperation")
    if not os.path.exists(BIN_INSTRUMENT):
        raise Exception("Please run 'make instrument' to generate target binary file.")


    for dir in args.dir:
        files = find_gotest_in_folder(dir)

        for f in files:
            abs_f = os.path.abspath(f)

            instrument_gotest(BIN_INSTRUMENT, abs_f, op_out)



if __name__ == "__main__":
    main()