package bmoc

import (
	"fmt"
	"github.com/system-pclub/GCatch/GCatch/analysis/pointer"
	"github.com/system-pclub/GCatch/GCatch/config"
	"github.com/system-pclub/GCatch/GCatch/instinfo"
	"github.com/system-pclub/GCatch/GCatch/syncgraph"
	"strconv"
)



func Detect() {
	stPtrResult, vecStOpValue := pointer.AnalyzeAllSyncOp()
	if stPtrResult == nil || vecStOpValue == nil {
		return
	}

	// When the pointer analysis has any uncertain alias relationship, report Not Sure.
	// This is in GCatch/analysis/pointer/utils.go, func mergeAlias()
	vecChannelOri := pointer.WithdrawAllChan(stPtrResult, vecStOpValue)

	vecLocker := pointer.WithdrawAllTraditionals(stPtrResult, vecStOpValue) // May delete

	mapDependency := syncgraph.GenDMap(vecChannelOri, vecLocker) // May delete

	vecChannel := []*instinfo.Channel{}
	for _, ch := range vecChannelOri {
		if OKToCheck(ch) == true { // some channels may come from SDK like testing. Ignore them
			vecChannel = append(vecChannel, ch)
		}
	}

	if len(vecChannel) == 0 { // Definitely no channel safety of liveness violations
		syncgraph.ReportNoViolation()
		return
	}



	// Check all channels together
	var boolFoundBug bool
	for _, ch := range vecChannel {
		boolFoundBug = boolFoundBug || CheckCh(ch, vecChannel, vecLocker, mapDependency)
	}

	if boolFoundBug {
		syncgraph.ReportViolation()
	} else {
		syncgraph.ReportNoViolation()
	}
}

var countCh int
var countUnbufferBug int
var countBufferBug int

func OKToCheck(ch *instinfo.Channel) (boolCheck bool) {
	boolCheck = false

	if ch.MakeInst == nil {
		return
	}
	pkg := ch.MakeInst.Parent().Pkg
	if pkg == nil {
		return
	}
	pkgOfPkg := pkg.Pkg
	if pkgOfPkg == nil {
		return
	}
	if config.IsPathIncluded(pkgOfPkg.Path()) == false {
		return
	}

	p := config.Prog.Fset.Position(ch.MakeInst.Pos())
	strChHash := ch.MakeInst.Parent().String() + ch.MakeInst.String() + ch.MakeInst.Name() + strconv.Itoa(p.Line)
	if _, checked := config.MapHashOfCheckedCh[strChHash]; checked {
		return
	}

	boolCheck = true
	config.MapHashOfCheckedCh[strChHash] = struct{}{}
	countCh++
	return
}

func CheckCh(ch *instinfo.Channel, vecChannel []*instinfo.Channel, vecLocker []*instinfo.Locker, mapDependency map[interface{}]*syncgraph.DPrim) (boolFoundBug bool) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	syncGraph, err := syncgraph.BuildGraph(ch, vecChannel, vecLocker, mapDependency)
	if err != nil { // Met some error
		if config.Print_Debug_Info {
			fmt.Println("-----count_ch:", countCh)
		}
		fmt.Println("Error when building graph", err.Error())
		syncgraph.ReportNotSure()
		return
	}

	syncGraph.ComputeFnOnOpPath()
	syncGraph.OptimizeBB_V1()

	syncGraph.SetEnumCfg(1, false, true)

	syncGraph.EnumerateAllPathCombinations()

	if ch.Buffer == instinfo.DynamicSize {
		// If this is a buffered channel with dynamic size and no critical section is found, skip this channel
		syncgraph.ReportNotSure()
	} else {
		boolFoundBug = syncGraph.CheckWithZ3()
	}
	return
}