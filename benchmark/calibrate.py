import sys
from typing import List, Tuple

def main():
    native_report_file = sys.argv[1]
    inst_report_file = sys.argv[2]

    natives = get_test_and_dur_from_report_file(native_report_file)
    insts = get_test_and_dur_from_report_file(inst_report_file)

    common_cnt = 0
    common_native_dur = 0
    common_inst_dur = 0
    for k, v in natives.items():
        if k in insts:
            common_cnt += 1
            common_native_dur += v
            common_inst_dur += insts[k]
    
    print(f"native: {len(natives)} inst: {len(insts)} common: {common_cnt}")
    print(f"native {common_native_dur/common_cnt:0.4f} seconds / test")
    print(f"inst {common_inst_dur/common_cnt:0.4f} seconds / test")




def get_test_and_dur_from_report_file(file:str):
    rec = {}
    with open(file, "r") as f:
        lines = f.read().splitlines()
        for line in lines:
            if line.find("->") != -1:
                idx = line.rfind(":")
                test_part = line[:idx]
                dur_part = line[idx+1:line.rfind(" ")]
                test = test_part[test_part.rfind("/")+1:]
                dur = float(dur_part)
                rec[test] = dur
    return rec


if __name__ == "__main__":
    main()