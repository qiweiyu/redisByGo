package core

import (
	"container/list"
	"runtime"
	"time"
)

const (
	dataNodeTypeHash      = 1
	dataNodeTypeList      = 2
	dataNodeTypeSet       = 3
	dataNodeTypeSortedSet = 4
	dataNodeTypeString    = 5
)

type dataNode struct {
	key         string
	dataType    int
	dataPointer interface{}
	deadTimer   *time.Timer
}

var dataNodeMap map[string]*dataNode
var lockSig chan int

func init() {
	lockSig = make(chan int, 1)
	dataNodeMap = make(map[string]*dataNode)
}

func lock(_ string) {
	lockSig <- 1
}

func unlock(_ string) {
	<-lockSig
}

func setToDb(key string, node *dataNode) {
	dataNodeMap[key] = node
}

func getFromDb(key string) (*dataNode, bool) {
	node, ex := dataNodeMap[key]
	return node, ex
}

func rmFromDb(key string) (*dataNode, bool) {
	node, ex := dataNodeMap[key]
	if ex {
		delete(dataNodeMap, key)
	}
	return node, ex
}

func flushDb() {
	dataNodeMap = make(map[string]*dataNode)
	runtime.GC()
}

func rmIfEmpty(key string) bool {
	node, ex := dataNodeMap[key]
	if ex == false {
		return false
	}
	switch node.dataType {
	case dataNodeTypeHash:
		if dataNode, ok := node.dataPointer.(hashNodeData); ok {
			if len(dataNode) > 0 {
				return false
			}
		}
	case dataNodeTypeList:
		if dataNode, ok := node.dataPointer.(*list.List); ok {
			if dataNode.Len() > 0 {
				return false
			}
		}
	case dataNodeTypeSet:
		if dataNode, ok := node.dataPointer.(setsNodeData); ok {
			if len(dataNode) > 0 {
				return false
			}
		}
	case dataNodeTypeString:
		if _, ok := node.dataPointer.(string); ok {
			return false
		}
	}
	delete(dataNodeMap, key)
	return true
}

func (this *dataNode) setTTL(ttlMs int) {
	if ttlMs > 0 {
		duration := time.Duration(ttlMs * 1e6)
		if this.deadTimer != nil {
			this.deadTimer.Stop()
		}
		this.deadTimer = time.AfterFunc(duration, func() {
			rmFromDb(this.key)
		})
	} else {
		if this.deadTimer != nil {
			this.deadTimer.Stop()
		}
		this.deadTimer = nil
	}
}
