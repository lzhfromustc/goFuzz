package util

import "github.com/system-pclub/GCatch/GCatch/tools/go/ssa"

func IsInstInVec(inst ssa.Instruction, vec []ssa.Instruction) bool {
	for _, elem := range vec {
		if elem == inst {
			return true
		}
	}
	return false
}

func VecFnForVecInst(vecInst []ssa.Instruction) []*ssa.Function {
	result := []*ssa.Function{}

	mapFn := make(map[*ssa.Function]struct{})
	for _,inst := range vecInst {
		mapFn[inst.Parent()] = struct{}{}
	}

	for fn,_ := range mapFn {
		result = append(result, fn)
	}

	return result
}