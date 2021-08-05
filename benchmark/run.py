#!/usr/bin/env python3
import shutil
import subprocess
import pathlib
import os
from shutil import copytree
from time import time


PROJ_ROOT_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
INST_SCRIPT = os.path.join(PROJ_ROOT_DIR, "goFuzz/scripts/instrument.py")
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
    """
    step 1: copy the `tests` folder and instrument that folder (in order to keep clean at `test` folder)
    step 2: compile native and instrumented package into binary file: native.test and inst.test
    step 3: run all tests 5 times and get average speeds regarding to native and inst.
    """
    print(f"project root: {PROJ_ROOT_DIR}")

    tests = ["TestOneSelect","TestCockroach1462"]
    
    # step 1
    copytree(TEST_PKG, TEST_PKG_INST_TMP)
    inst_dir(TEST_PKG_INST_TMP)

    # step 2
    compile_test_bin(TEST_PKG_INST_TMP, TEST_BIN_INST)
    compile_test_bin(TEST_PKG, TEST_BIN_NATIVE)

    # step 3
    total_native_dur = 0
    total_inst_dur = 0
    std_input_content = ""
    with open(STD_INPUT_FILE, "r") as f:
        std_input_content = f.read()

    inst_run_env = os.environ.copy()
    inst_run_env["GF_INPUT_FILE"] = STD_INPUT_FILE
    inst_run_env["GF_RECORD_FILE"] = STD_RECORD_FILE
    inst_run_env["GF_OUTPUT_FILE"] = STD_OUTPUT_FILE

    for t in tests:
        native_dur = benchmark(10, lambda: subprocess.run([TEST_BIN_NATIVE, "-test.v", "-test.run", t]))
        total_native_dur += native_dur
        inst_dur = benchmark(10, lambda: subprocess.run(
            [TEST_BIN_INST, "-test.run", t],
            env=inst_run_env
        ), lambda: restore_inst_run(std_input_content)
            
        )
        total_inst_dur += inst_dur
        print(f"{t}: native {native_dur:.04f} seconds")
        print(f"{t}: inst {inst_dur:.04f} seconds")
    
    print(f"total native average {total_native_dur/len(tests):.04f} seconds / test")
    print(f"total inst average {total_inst_dur/len(tests):.04f} seconds / test")

def inst_dir(dir: str):
    subprocess.run([INST_SCRIPT, dir]).check_returncode()

def compile_test_bin(pkg_dir:str, dest:str):
    subprocess.run(
        ["go","test","-c", "-o", dest, pkg_dir], 
        cwd=BENCHMARK_DIR
    ).check_returncode()

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