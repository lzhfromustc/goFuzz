#!/usr/bin/env python3
import shutil
import subprocess
import pathlib
import os
import argparse
from shutil import copytree
from time import time


PROJ_ROOT_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
INST_SCRIPT = os.path.join(PROJ_ROOT_DIR, "goFuzz/scripts/instrument.py")
PATCH_RUNTIME_SCRIPT = os.path.join(PROJ_ROOT_DIR, "scripts/patch-go-runtime.sh")

BENCHMARK_DIR = os.path.join(PROJ_ROOT_DIR, "benchmark")
STD_INPUT_FILE = os.path.join(BENCHMARK_DIR, "std-input")

TEST_PKG = os.path.join(BENCHMARK_DIR, "tests")


TMP_FOLDER = os.path.join(PROJ_ROOT_DIR, "benchmark/tmp")
TEST_PKG_INST_TMP = os.path.join(TMP_FOLDER, "instrument")
TEST_BIN_NATIVE = os.path.join(TMP_FOLDER, "native.test")
TEST_BIN_INST = os.path.join(TMP_FOLDER, "inst.test")

STD_RECORD_FILE = os.path.join(TMP_FOLDER, "record")
STD_OUTPUT_FILE = os.path.join(TMP_FOLDER, "output")



def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('benchmark_type', choices=["native", "inst"])

    args = parser.parse_args()

    print(f"project root: {PROJ_ROOT_DIR}")
    tests = ["TestOneSelect","TestCockroach1462","TestNoSelect","TestEtcd6873"]
    
    # prepare test bin
    if args.benchmark_type == "inst":
        patch_go_runtime()
        copytree(TEST_PKG, TEST_PKG_INST_TMP)
        inst_dir(TEST_PKG_INST_TMP)
        compile_test_bin(TEST_PKG_INST_TMP, TEST_BIN_INST)
    else:
        compile_test_bin(TEST_PKG, TEST_BIN_NATIVE)

    
    # run tests
    total_dur = 0

    if args.benchmark_type == "inst":
        inst_run_env = os.environ.copy()
        inst_run_env["GF_BENCHMARK"] = "1"

    for t in tests:

        if args.benchmark_type == "inst":
            dur = benchmark(10, lambda: subprocess.run(
                [TEST_BIN_INST, "-test.run", t], env=inst_run_env))
        else:
            dur = benchmark(10, lambda: subprocess.run([TEST_BIN_NATIVE, "-test.v", "-test.run", t]))
        
        print(f"{t}: {dur:.04f} seconds")
        total_dur += dur

    print(f"total average {total_dur/len(tests):.04f} seconds / test")

def inst_dir(dir: str):
    subprocess.run([INST_SCRIPT, dir]).check_returncode()

def compile_test_bin(pkg_dir:str, dest:str):
    subprocess.run(
        ["go","test","-c", "-o", dest, pkg_dir], 
        cwd=BENCHMARK_DIR
    ).check_returncode()

def patch_go_runtime():
    subprocess.run([PATCH_RUNTIME_SCRIPT]).check_returncode()

def restore_inst_run(std_input_content:str):
    if os.path.exists(STD_RECORD_FILE):
        os.remove(STD_RECORD_FILE)
    
    if os.path.exists(STD_OUTPUT_FILE):
        os.remove(STD_OUTPUT_FILE)

    with open(STD_INPUT_FILE, 'w') as f:
        f.write(std_input_content)

def benchmark(reps, func, prefunc=None):
    dur = 0
    for _ in range(0, reps):
        if prefunc:
            prefunc()
        start = time()
        func()
        end = time()
        dur += (end-start)
    return dur / reps

if __name__ == "__main__":
    main()