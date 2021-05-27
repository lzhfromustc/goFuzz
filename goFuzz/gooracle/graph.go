package gooracle

import (
	"github.com/yourbasic/graph"
)

// HasCyclicDependency is trying to find any cyclic-wait state in a list of blocked goroutine
// WIP
func HasCyclicDependency(blockedGs []*GoInfo) bool {
	var allRefGs = make([]*GoInfo, 0, len(blockedGs))

	var allRefChs = []*ChanInfo{}

	var ch2Idx map[*ChanInfo]int = make(map[*ChanInfo]int)
	var g2Idx map[*GoInfo]int = make(map[*GoInfo]int)

	for _, bg := range blockedGs {
		bg.mapChanInfo.Range(func(key, value interface{}) bool {
			chInfo, _ := key.(*ChanInfo)
			allRefChs = append(allRefChs, chInfo)
			return true
		})
	}

	for idx, ch := range allRefChs {
		ch2Idx[ch] = idx
	}

	for idx, g := range allRefGs {
		g2Idx[g] = idx
	}

	g := graph.New(len(allRefChs) + len(allRefGs))

	for _, bg := range blockedGs {

		bgIdx := g2Idx[bg]
		bg.mapChanInfo.Range(func(key, value interface{}) bool {
			chInfo, _ := key.(*ChanInfo)
			chIdx := ch2Idx[chInfo]
			g.Add(bgIdx, chIdx)
			return true
		})

		for blockedByCh, _ := range bg.BlockMap {
			bChIdx := ch2Idx[blockedByCh]
			g.Add(bChIdx, bgIdx)
		}

	}

	return graph.Acyclic(g)
}
