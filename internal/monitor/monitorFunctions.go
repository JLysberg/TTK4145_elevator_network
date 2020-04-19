package monitor

import . "../common/types"

var getQueueCopy = make(chan []FloorState)
var getGlobalCopy = make(chan GlobalInfo)
var remOrders = make(chan remOrdersParams)

type remOrdersParams struct {
	Floor int
	ID    int
}

func createQueueCopy(queue []FloorState) []FloorState {
	copy := make([]FloorState, len(queue))
	for i, k := range queue {
		copy[i] = k
	}
	return copy
}

func createGlobalCopy(global GlobalInfo) GlobalInfo {
	copy := global
	copy.Orders = make([][]FloorState, len(global.Orders))
	copy.Nodes = make([]LocalInfo, len(global.Nodes))
	for i, v := range global.Orders {
		copy.Orders[i] = make([]FloorState, len(v))
		for j, k := range v {
			copy.Orders[i][j] = k
		}
	}
	for i, v := range global.Nodes {
		copy.Nodes[i] = v
	}
	
	return copy
}

func equalOrderMatrix(m1 [][]FloorState, m2 [][]FloorState) bool {
	if len(m1) == len(m2) && len(m1[0]) == len(m2[0]) {
		for i, v := range m1 {
			for j, k := range v {
				if m2[i][j] != k {
					return false
				}
			}
		}
	} else {
		return false
	}
	return true
}

func equalNodeArray(a1 []LocalInfo, a2 []LocalInfo) bool {
	if len(a1) == len(a2) {
		for i, v := range a1 {
			if a2[i] != v {
				return false
			}
		}
	} else {
		return false
	}
	return true
}