package core

import (
	"strconv"
)

type hashNodeData map[string]string

func createHashNode(key string) *dataNode {
	var node = new(dataNode)
	node.key = key
	node.dataType = dataNodeTypeHash
	node.dataPointer = nil
	node.setTTL(0)
	return node
}

func baseHGetAll(key string) (hashNodeData, *cmdResult) {
	data, ex := getFromDb(key)
	if ex {
		if data.dataType != dataNodeTypeHash {
			return nil, commandResErrType()
		}
		if hashNode, ok := data.dataPointer.(hashNodeData); ok {
			return hashNode, nil
		}
		return nil, commandResErrType()
	} else {
		return nil, commandResEmptyArray()
	}
}

func baseHSet(key, field, value string) *cmdResult {
	data, ex := getFromDb(key)
	var hashNode hashNodeData
	if ex {
		if data.dataType != dataNodeTypeHash {
			return commandResErrType()
		}
		var ok bool
		if hashNode, ok = data.dataPointer.(hashNodeData); ok == false {
			hashNode = make(hashNodeData)
		}
	} else {
		data = createHashNode(key)
		hashNode = make(hashNodeData)
	}
	_, exField := hashNode[field]
	hashNode[field] = value
	data.dataPointer = interface{}(hashNode)
	setToDb(key, data)
	if exField {
		return commandResInt(0)
	}
	return commandResInt(1)
}

func baseHIncr(key, field, value string) *cmdResult {
	delta, err := strconv.Atoi(value)
	if err != nil {
		return commandResErrParseInt("value")
	}
	hashNode, cmd := baseHGetAll(key)
	var valueInt int
	if hashNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		valueInt = 0
	} else if valueStr, ex := hashNode[field]; ex {
		var err error
		valueInt, err = strconv.Atoi(valueStr)
		if err != nil {
			return commandResErrParseInt("value")
		}
	} else {
		valueInt = 0
	}
	valueInt = valueInt + delta
	baseHSet(key, field, strconv.Itoa(valueInt))
	return commandResInt(valueInt)
}

func doHDel(opt ...string) *cmdResult {
	key := opt[0]
	dataNode, cmd := baseHGetAll(key)
	if dataNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	var count = 0
	for _, field := range opt[1:] {
		if _, ex := dataNode[field]; ex {
			delete(dataNode, field)
			count++
		}
	}
	rmIfEmpty(key)
	return commandResInt(count)
}

func doHExists(opt ...string) *cmdResult {
	key := opt[0]
	field := opt[1]
	dataNode, cmd := baseHGetAll(key)
	if dataNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	if _, ex := dataNode[field]; ex {
		return commandResInt(1)
	}
	return commandResInt(0)
}

func doHGet(opt ...string) *cmdResult {
	key := opt[0]
	field := opt[1]
	dataNode, cmd := baseHGetAll(key)
	if dataNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResNil()
	}
	if data, ex := dataNode[field]; ex {
		return commandResString(data)
	}
	return commandResNil()
}

func doHGetAll(opt ...string) *cmdResult {
	key := opt[0]
	dataNode, cmd := baseHGetAll(key)
	if cmd != nil {
		return cmd
	}
	cmdList := make([]*cmdResult, len(dataNode)*2)
	var i = 0
	for field, value := range dataNode {
		cmdList[i] = commandResString(field)
		i++
		cmdList[i] = commandResString(value)
		i++
	}
	return commandResArray(cmdList)
}

func doHIncrBy(opt ...string) *cmdResult {
	return baseHIncr(opt[0], opt[1], opt[2])
}

func doHIncrByFloat(opt ...string) *cmdResult {
	key := opt[0]
	field := opt[1]
	valueString := opt[2]
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return commandResErr("ERR value is not a valid float")
	}
	hashNode, cmd := baseHGetAll(key)
	var valueFloat float64
	if hashNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		valueFloat = 0
	} else if valueStr, ex := hashNode[field]; ex {
		var err error
		valueFloat, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return commandResErr("ERR value is not a valid float")
		}
	} else {
		valueFloat = 0
	}
	valueFloat = valueFloat + value
	baseHSet(key, field, strconv.FormatFloat(valueFloat, 'f', -1, 64))
	return commandResString(strconv.FormatFloat(valueFloat, 'f', -1, 64))
}

func doHKeys(opt ...string) *cmdResult {
	key := opt[0]
	dataNode, cmd := baseHGetAll(key)
	if cmd != nil {
		return cmd
	}
	cmdList := make([]*cmdResult, len(dataNode))
	var i = 0
	for field := range dataNode {
		cmdList[i] = commandResString(field)
		i++
	}
	return commandResArray(cmdList)
}

func doHLen(opt ...string) *cmdResult {
	key := opt[0]
	dataNode, cmd := baseHGetAll(key)
	if cmd != nil {
		if cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	return commandResInt(len(dataNode))
}

func doHMGet(opt ...string) *cmdResult {
	key := opt[0]
	resList := make([]*cmdResult, len(opt)-1)
	dataNode, cmd := baseHGetAll(key)
	if dataNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		dataNode = make(hashNodeData)
	}
	for i, field := range opt[1:] {
		if str, ex := dataNode[field]; ex {
			resList[i] = commandResString(str)
		} else {
			resList[i] = commandResNil()
		}
	}
	return commandResArray(resList)
}

func doHMSet(opt ...string) *cmdResult {
	if len(opt)%2 == 0 {
		return commandResErr("ERR wrong number of arguments for HMSET")
	}
	key := opt[0]
	dataNode, cmd := baseHGetAll(key)
	if dataNode == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
	}
	for i := 1; i < len(opt); i += 2 {
		baseHSet(key, opt[i], opt[i+1])
	}
	return commandResOk()
}

func doHSet(opt ...string) *cmdResult {
	key := opt[0]
	field := opt[1]
	value := opt[2]
	return baseHSet(key, field, value)
}

func doHSetNx(opt ...string) *cmdResult {
	cmd := doHExists(opt...)
	if cmd.resType != resTypeInt {
		return cmd
	}
	if cmd.resInt == 1 {
		return commandResInt(0)
	}
	return doHSet(opt...)
}

func doHStrlen(opt ...string) *cmdResult {
	cmd := doHGet(opt...)
	if cmd.resType == resTypeString {
		return commandResInt(len(cmd.resMsg))
	} else if cmd.resType == resTypeNil {
		return commandResInt(0)
	}
	return cmd
}

func doHVals(opt ...string) *cmdResult {
	key := opt[0]
	dataNode, cmd := baseHGetAll(key)
	if cmd != nil {
		return cmd
	}
	cmdList := make([]*cmdResult, len(dataNode))
	var i = 0
	for _, value := range dataNode {
		cmdList[i] = commandResString(value)
		i++
	}
	return commandResArray(cmdList)
}
