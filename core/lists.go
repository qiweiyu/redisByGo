package core

import (
	"container/list"
	"strconv"
	"strings"
)

func createListNode(key string, l *list.List) *dataNode {
	var node = new(dataNode)
	node.key = key
	node.dataType = dataNodeTypeList
	node.dataPointer = interface{}(l)
	node.setTTL(0)
	return node
}

func getStringFromElement(e *list.Element) string {
	if s, ok := e.Value.(string); ok {
		return s
	}
	return ""
}

func baseLGet(key string) (*list.List, *cmdResult) {
	data, ex := getFromDb(key)
	if ex {
		if data.dataType != dataNodeTypeList {
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

func baseLSet(key string) *list.List {
	l := list.New()
	node := createListNode(key, l)
	setToDb(key, node)
	return l
}

func doLIndex(opt ...string) *cmdResult {
	key := opt[0]
	indexStr := opt[1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	l, cmd := baseLGet(key)
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResNil()
		}
	}
	len := l.Len()
	if index >= len {
		return commandResNil()
	}
	if index < 0 {
		index = index + len
	}
	if index < 0 {
		return commandResNil()
	}
	e := l.Front()
	for i := 0; i < index; i++ {
		e = e.Next()
	}
	return commandResString(getStringFromElement(e))
}

func doLInsert(opt ...string) *cmdResult {
	key := opt[0]
	pos := strings.ToLower(opt[1])
	pivot := opt[2]
	value := opt[3]
	if pos != "before" && pos != "after" {
		return commandResErrSyntax()
	}
	l, cmd := baseLGet(key)
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResInt(0)
		}
	}
	e := l.Front()
	found := false
	for found == false && e != nil {
		if getStringFromElement(e) == pivot {
			found = true
			break
		}
		e = e.Next()
	}
	if found {
		if pos == "before" {
			l.InsertBefore(interface{}(value), e)
		} else {
			l.InsertAfter(interface{}(value), e)
		}
		return commandResInt(l.Len())
	} else {
		return commandResInt(-1)
	}
}

func doLLen(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if cmd != nil {
		if cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResInt(0)
		}
	}
	return commandResInt(l.Len())
}

func doLPop(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if cmd != nil {
		return cmd
	}
	e := l.Front()
	if e == nil {
		return commandResNil()
	}
	s := getStringFromElement(e)
	l.Remove(e)
	rmIfEmpty(opt[0])
	return commandResString(s)
}

func doLPush(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			l = baseLSet(opt[0])
		}
	}
	for _, v := range opt[1:] {
		l.PushFront(interface{}(v))
	}
	return commandResInt(l.Len())
}

func doLPushX(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResInt(0)
		}
	}
	l.PushFront(interface{}(opt[1]))
	return commandResInt(l.Len())
}

func doLRange(opt ...string) *cmdResult {
	key := opt[0]
	startStr := opt[1]
	endStr := opt[2]
	var start, end int
	var err error
	start, err = strconv.Atoi(startStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	end, err = strconv.Atoi(endStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	l, cmd := baseLGet(key)
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResEmptyArray()
		}
	}
	len := l.Len()
	if start < 0 {
		start = len + start
	}
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len + end
	}
	if end >= len {
		end = len - 1
	}
	if start > end {
		return commandResEmptyArray()
	}
	resList := make([]*cmdResult, end+1-start)
	e := l.Front()
	for i := 0; i < start; i++ {
		e = e.Next()
	}
	j := 0
	for i := start; i <= end; i++ {
		resList[j] = commandResString(getStringFromElement(e))
		j++
		e = e.Next()
	}
	return commandResArray(resList)
}

func doLRem(opt ...string) *cmdResult {
	key := opt[0]
	countStr := opt[1]
	value := opt[2]
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	l, cmd := baseLGet(key)
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResInt(0)
		}
	}
	totalRm := 0
	switch {
	case count > 0:
		e := l.Front()
		for {
			if totalRm >= count || e == nil {
				break
			}
			s := getStringFromElement(e)
			if s == value {
				rmE := e
				totalRm++
				e = e.Next()
				l.Remove(rmE)
			} else {
				e = e.Next()
			}
		}
	case count == 0:
		e := l.Front()
		for {
			if e == nil {
				break
			}
			s := getStringFromElement(e)
			if s == value {
				rmE := e
				totalRm++
				e = e.Next()
				l.Remove(rmE)
			} else {
				e = e.Next()
			}
		}
	case count < 0:
		e := l.Back()
		for {
			if totalRm >= count*-1 || e == nil {
				break
			}
			s := getStringFromElement(e)
			if s == value {
				rmE := e
				totalRm++
				e = e.Prev()
				l.Remove(rmE)
			} else {
				e = e.Prev()
			}
		}
	}
	return commandResInt(totalRm)
}

func doLSet(opt ...string) *cmdResult {
	key := opt[0]
	indexStr := opt[1]
	value := opt[2]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	l, cmd := baseLGet(key)
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResErr("ERR no such key")
		}
	}
	len := l.Len()
	if index >= len {
		return commandResErr("ERR index out of range")
	}
	if index < 0 {
		index = index + len
	}
	if index < 0 {
		return commandResErr("ERR index out of range")
	}
	e := l.Front()
	for i := 0; i < index; i++ {
		e = e.Next()
	}
	e.Value = interface{}(value)
	return commandResOk()
}

func doLTrim(opt ...string) *cmdResult {
	key := opt[0]
	startStr := opt[1]
	endStr := opt[2]
	var start, end int
	var err error
	start, err = strconv.Atoi(startStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	end, err = strconv.Atoi(endStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	l, cmd := baseLGet(key)
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResOk()
		}
	}
	len := l.Len()
	if start < 0 {
		start = len + start
	}
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len + end
	}
	if end > len {
		end = len - 1
	}
	if start > end {
		rmFromDb(key)
	}
	for i := 0; i < start; i++ {
		e := l.Front()
		l.Remove(e)
	}
	if end < len-1 {
		for i := end + 1; i < len; i++ {
			e := l.Back()
			l.Remove(e)
		}
	}
	return commandResOk()
}

func doRPop(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if cmd != nil {
		return cmd
	}
	e := l.Back()
	if e == nil {
		return commandResNil()
	}
	s := getStringFromElement(e)
	l.Remove(e)
	rmIfEmpty(opt[0])
	return commandResString(s)
}

func doRPopLPush(opt ...string) *cmdResult {
	l1, cmd := baseLGet(opt[0])
	if cmd != nil {
		return cmd
	}
	e := l1.Back()
	if e == nil {
		return commandResNil()
	}
	s := getStringFromElement(e)
	var l2 *list.List
	l2, cmd = baseLGet(opt[1])
	if l2 == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			l2 = baseLSet(opt[1])
		}
	}
	l2.PushBack(interface{}(s))
	l1.Remove(e)
	rmIfEmpty(opt[0])
	return commandResString(s)
}

func doRPush(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			l = baseLSet(opt[0])
		}
	}
	for _, v := range opt[1:] {
		l.PushBack(interface{}(v))
	}
	return commandResInt(l.Len())
}

func doRPushX(opt ...string) *cmdResult {
	l, cmd := baseLGet(opt[0])
	if l == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		} else {
			return commandResInt(0)
		}
	}
	l.PushBack(interface{}(opt[1]))
	return commandResInt(l.Len())
}
