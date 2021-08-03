#!/usr/bin/env python3
import argparse
import glob
import subprocess
from typing import List
import pathlib
import os
from concurrent.futures import ThreadPoolExecutor

PROJ_ROOT_DIR = pathlib.Path(os.path.realpath(__file__)).parent.parent.as_posix()
BIN_INSTRUMENT = None
BIN_PRINTOPERATION = None

# Metrics
NUM_OF_FILES_INSTRUMENTED = 0
NUM_OF_SELECTS = 0

def find_go_files_in_folder(folder: str) -> List[str]:
    """Find all *_test.go files under the given folder

    Args:
        folder (str): A foder path

    Returns:
        [str]: A list of file path
    """
    if folder.endswith("/"):
        folder = folder[:-1]
    glob_path = os.path.join(folder, '**', '*.go')
    return glob.glob(glob_path, recursive=True)

def generate_ch_statistics(bin:str, go_file:str, output_file:str):
    """Dump channel statistics to the given output file by bin/printOperation
    :param file:
    :param output:
    :return:
    """
    res = subprocess.run([bin, f"-file={go_file}", f"-output={output_file}"])
    res.check_returncode()

def instrument_go_file(bin:str, go_file: str, rec_out_file:str):
    """Instrument Golang test file by bin/instrument

    Args:
        file (str): Golang test file going to be instrumented
    """
    res = subprocess.run([bin, f"-file={go_file}", f"-output={rec_out_file}"], stdout=subprocess.PIPE)
    res.check_returncode()
    num_of_selects = parse_inst_output(res.stdout.decode('utf-8'))
    global NUM_OF_FILES_INSTRUMENTED, NUM_OF_SELECTS
    NUM_OF_FILES_INSTRUMENTED += 1
    NUM_OF_SELECTS += num_of_selects


def parse_inst_output(output: str) -> int:
    """Parse stdout of binary `instrument` to get number of selects
       for that given instrumented go file.

    Args:
        output (str): stdout string

    Returns:
        int: number of selects
    """
    lines = output.splitlines()
    for line in lines:
        if line.find("number of selects: ") != -1:
            num_str = line[line.rfind(" ")+1:]
            return int(num_str)

    return 0

def main():
    parser = argparse.ArgumentParser(description='Instrument Golang test files.')
    parser.add_argument("dir", metavar='DIR', type=str, nargs='+',
                        help='directory contains *_test.go files')
    parser.add_argument("--op-out", help="Output path of premitives statistics(of Go source code in the given dir) ")

    args = parser.parse_args()

    op_out = args.op_out
    if not op_out:
        # If output is not given, create a file under CWD.
        op_out = "op-out"
    op_out = os.path.abspath(op_out)

    global PROJ_ROOT_DIR
    global BIN_INSTRUMENT
    global BIN_PRINTOPERATION

    BIN_INSTRUMENT = os.path.join(PROJ_ROOT_DIR, "bin/instrument")
    BIN_PRINTOPERATION = os.path.join(PROJ_ROOT_DIR, "bin/printOperation")
    if not os.path.exists(BIN_INSTRUMENT):
        raise Exception("Please run 'make instrument' to generate target binary file.")

    all_go_files = []
    for d in args.dir:
        files = find_go_files_in_folder(d)
        all_go_files.extend(files)

    futures = []
    with ThreadPoolExecutor(5) as executor:
        for f in all_go_files:
            future = executor.submit(instrument_go_file, BIN_INSTRUMENT, f, op_out)
            futures.append(future)

    for f in futures:
        f.result()
    
    print(f"total number of instrumented files: {NUM_OF_FILES_INSTRUMENTED}")
    print(f"total number of selects: {NUM_OF_SELECTS}")






if __name__ == "__main__":
    main()