import os
from enum import Enum
import time
import datetime
from typing import List
import matplotlib.pyplot as plt
from matplotlib.ticker import PercentFormatter
from matplotlib import ticker

total_worker_num: int = 5

class STAGE(Enum):
    init = 0
    deter = 1
    rand = 2

class STAGE_INFO():
    stage:STAGE
    total_runtime: int = 0
    total_run_num: int = 0
    total_found_bug: int = 0
    total_found_uniq_bug: int = 0
    def __init__(self, stage_in) -> None:
        self.stage = stage_in

class WORKER:
    id:int = 0
    stage:STAGE = STAGE.init
    start_unixtime:int = 0


file_fd = open("./fuzzer.log", "r")

lines = file_fd.readlines()

worker_list = []

make_chan_time = []
make_chan_cov = []
chan_op_time = []
chan_op_cov = []
other_op_time = []
other_op_cov = []

for i in range(total_worker_num):
    worker = WORKER()
    worker.id = i
    worker_list.append(worker)

init_stage = STAGE_INFO(STAGE.init)
deter_stage = STAGE_INFO(STAGE.deter)
rand_stage = STAGE_INFO(STAGE.rand)

fuzzing_start_unix_time:int = 0

for cur_line in lines:
    if fuzzing_start_unix_time == 0:
        cur_line_split = cur_line.split(" ")
        time_str = cur_line_split[0] + " " + cur_line_split [1]
        fuzzing_start_unix_time = int(datetime.datetime.strptime(time_str, "%Y/%m/%d %H:%M:%S").timestamp())

    if "Working on" in cur_line: # This is the start of the task
        # get stage
        cur_line_split = cur_line.split(" ")
        task_name = cur_line_split[-1]
        cur_stage = STAGE.init
        if "init" in task_name:
            cur_stage = STAGE.init
        elif "deter" in task_name:
            cur_stage = STAGE.deter
        elif "rand" in task_name:
            cur_stage = STAGE.rand
        else:
            os.error("Unexpected logic error. No Stage info detected in task_name")

        # get unixtime
        time_str = cur_line_split[0] + " " + cur_line_split [1]
        start_time_int = int(datetime.datetime.strptime(time_str, "%Y/%m/%d %H:%M:%S").timestamp())

        # get worker num
        cur_worker_num = (cur_line.split("]")[0]).split(" ")[-1]
        cur_worker_num = int(cur_worker_num)

        # Retrive worker previous task runtime. 
        cur_worker_info:WORKER = worker_list[cur_worker_num]
        prev_task_runtime = 0
        if cur_worker_info.start_unixtime != 0:
            prev_task_runtime = start_time_int - cur_worker_info.start_unixtime
        cur_worker_info.start_unixtime = start_time_int

        prev_task_stage = cur_worker_info.stage
        if prev_task_stage == STAGE.init:
            init_stage.total_runtime += prev_task_runtime
            init_stage.total_run_num += 1
        elif prev_task_stage == STAGE.deter:
            deter_stage.total_runtime += prev_task_runtime
            deter_stage.total_run_num += 1
        elif prev_task_stage == STAGE.rand:
            rand_stage.total_runtime += prev_task_runtime
            rand_stage.total_run_num += 1
        else:
            os.error("Unexpected logic error. No Stage info detected in prev_task_stage")
        cur_worker_info.stage = cur_stage


    # take care of bug reports    
    elif "bug(s)" in cur_line:
        bug_num = int((cur_line.split(" bug(s)")[0]).split(" ")[-1])
        uniq_bug_num = int((cur_line.split(" unique bug(s)")[0]).split(" ")[-1])

        if "init" in cur_line:
            init_stage.total_found_bug += bug_num
            init_stage.total_found_uniq_bug += uniq_bug_num
        elif "deter" in cur_line:
            deter_stage.total_found_bug += bug_num
            deter_stage.total_found_uniq_bug += uniq_bug_num
        elif "rand" in cur_line:
            rand_stage.total_found_bug += bug_num
            rand_stage.total_found_uniq_bug += uniq_bug_num
        else:
            os.error("Unexpected logic error. No Stage info detected in cur_line")

    elif "cumulative channel make count" in cur_line:
        time_str = cur_line_split[0] + " " + cur_line_split [1]
        cur_time_int = int(datetime.datetime.strptime(time_str, "%Y/%m/%d %H:%M:%S").timestamp()) - fuzzing_start_unix_time
        cur_time_int = cur_time_int / 3600

        cov_str = (cur_line.split("coverage ")[1])[:-2]
        cov = float(cov_str)

        make_chan_time.append(cur_time_int)
        make_chan_cov.append(cov)

    elif "cumulative channel op count" in cur_line:
        time_str = cur_line_split[0] + " " + cur_line_split [1]
        cur_time_int = int(datetime.datetime.strptime(time_str, "%Y/%m/%d %H:%M:%S").timestamp()) - fuzzing_start_unix_time
        cur_time_int = cur_time_int / 3600

        cov_str = (cur_line.split("coverage ")[1])[:-2]
        cov = float(cov_str)

        chan_op_time.append(cur_time_int)
        chan_op_cov.append(cov)

    elif "cumulative other primitive op count" in cur_line:
        time_str = cur_line_split[0] + " " + cur_line_split [1]
        cur_time_int = int(datetime.datetime.strptime(time_str, "%Y/%m/%d %H:%M:%S").timestamp()) - fuzzing_start_unix_time
        cur_time_int = cur_time_int / 3600

        cov_str = (cur_line.split("coverage ")[1])[:-2]
        cov = float(cov_str)

        other_op_time.append(cur_time_int)
        other_op_cov.append(cov)


print("In total, we have init: run_num: %d, run_time: %d, total_bugs: %d, total_unique_bugs: %d" \
    % (init_stage.total_run_num, init_stage.total_runtime, init_stage.total_found_bug, init_stage.total_found_uniq_bug))

print("In total, we have deter: run_num: %d, run_time: %d, total_bugs: %d, total_unique_bugs: %d" \
    % (deter_stage.total_run_num, deter_stage.total_runtime, deter_stage.total_found_bug, deter_stage.total_found_uniq_bug))

print("In total, we have rand: run_num: %d, run_time: %d, total_bugs: %d, total_unique_bugs: %d" \
    % (rand_stage.total_run_num, rand_stage.total_runtime, rand_stage.total_found_bug, rand_stage.total_found_uniq_bug))

plt.figure()
ax = plt.subplot()

ax.plot(make_chan_time, make_chan_cov)
ax.plot(chan_op_time, chan_op_cov)
ax.plot(other_op_time, other_op_cov)

ax.yaxis.set_major_formatter(PercentFormatter())

plt.xlabel("Time (h)", fontsize = 20)
plt.ylabel('Coverage', fontsize = 20)
plt.legend(['make_chan_cov', 'chan_op_cov', 'other_op_cov'])
plt.xticks(fontsize=20)
plt.yticks(fontsize=20)
plt.tight_layout()
plt.savefig('gofuzz_cov.png', dpi = 200)
