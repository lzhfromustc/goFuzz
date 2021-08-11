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

    buckets = {}
    for k, v in natives.items():
        if k in insts:
            common_cnt += 1
            common_native_dur += v
            common_inst_dur += insts[k]
            inc = insts[k]/v - 1
            print(f"native {v:0.4f}, inst {insts[k]:0.4f}, inc: {inc:0.4f}")
            bucket = get_bucket_from_diff(inc)
            if bucket not in buckets:
                buckets[bucket] = 1
            else:
                buckets[bucket] += 1
    
    print(f"native: {len(natives)} inst: {len(insts)} common: {common_cnt}")
    print(f"native {common_native_dur:0.4f} seconds")
    print(f"inst {common_inst_dur:0.4f} seconds")
    print(f"native {common_native_dur/common_cnt:0.4f} seconds / test")
    print(f"inst {common_inst_dur/common_cnt:0.4f} seconds / test")
    for k, v in buckets.items():
        print(f"{k}: {v}")


def get_bucket_from_diff(diff: float)->str:
    if diff <= 0.1:
        return "10%"
    elif 0.1 < diff <= 0.3:
        return "10%-30%"
    elif 0.3 < diff <= 1:
        return "30%-100%"
    elif 1 < diff <= 3:
        return "100%-300%"
    elif 3 < diff :
        return ">300%"
    else:
        return "unknown"


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