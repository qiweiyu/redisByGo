package core

import (
	"container/list"
)

type sortedSetNode struct {
	score  float64
	member string
}

func createSSetList(key string, s *list.List) *dataNode {
	var node = new(dataNode)
	node.key = key
	node.dataType = dataNodeTypeSortedSet
	node.dataPointer = interface{}(s)
	node.setTTL(0)
	return node
}

func getSSetNodeFromElement(e *list.Element) *sortedSetNode {
	if s, ok := e.Value.(*sortedSetNode); ok {
		return s
	}
	return nil
}

func baseSSetListGet(key string) (*list.List, *cmdResult) {
	data, ex := getFromDb(key)
	if ex {
		if data.dataType != dataNodeTypeSortedSet {
			return nil, commandResErrType()
		}
		if l, ok := data.dataPointer.(*list.List); ok {
			return l, nil
		}
		return nil, commandResErrType()
	} else {
		return nil, commandResNil()
	}
}

func baseSSetListSet(key string) *list.List {
	l := list.New()
	node := createListNode(key, l)
	setToDb(key, node)
	return l
}

func doZAdd(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZCard(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZCount(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZIncrBy(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZInterStore(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZLexCount(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRange(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRangeByLex(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRevRangeByLex(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRangeByScore(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRank(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRem(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRemRrangeByLex(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRemRangeByRank(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRemRangeByScore(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRevRange(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRevRangeByScore(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZRevRank(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZScore(_ ...string) *cmdResult {
	return commandResString("todo")
}

func doZUnionStore(_ ...string) *cmdResult {
	return commandResString("todo")
}
