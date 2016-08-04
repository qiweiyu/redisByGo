package core

import (
	"strconv"
	"strings"
)

func createStringNode(key, str string, ttlOfMs int) *dataNode {
	var node = new(dataNode)
	node.key = key
	node.dataType = dataNodeTypeString
	node.dataPointer = interface{}(str)
	node.setTTL(ttlOfMs)
	return node
}

func baseGet(key string) *cmdResult {
	data, ex := getFromDb(key)
	if ex {
		if data.dataType != dataNodeTypeString {
			return commandResErrType()
		}
		if str, ok := data.dataPointer.(string); ok {
			return commandResString(str)
		} else {
			return commandResString("")
		}
	} else {
		return commandResNil()
	}
}

func baseSet(key, value string, ttlSetFlag bool, ttlMs int, nxFlag bool, xxFlag bool) *cmdResult {
	if ttlMs <= 0 && ttlSetFlag {
		return commandResErr("ERR invalid expire time")
	}
	data, ex := getFromDb(key)
	if ex {
		if nxFlag {
			return commandResNil()
		}
		data.dataPointer = interface{}(value)
		if ttlSetFlag {
			data.setTTL(ttlMs)
		}
	} else {
		if xxFlag {
			return commandResNil()
		}
		if ttlSetFlag == false {
			ttlMs = 0
		}
		data = createStringNode(key, value, ttlMs)
	}
	data.dataType = dataNodeTypeString
	setToDb(key, data)
	return commandResOk()
}

func baseMSet(nxFlag bool, opt ...string) *cmdResult {
	if len(opt)%2 == 1 {
		return commandResErr("ERR wrong number of arguments for MSET")
	}
	if nxFlag {
		for i := 0; i < len(opt); i += 2 {
			if _, ex := getFromDb(opt[i]); ex {
				return commandResInt(0)
			}
		}
	}
	for i := 0; i < len(opt); i += 2 {
		baseSet(opt[i], opt[i+1], false, 0, false, false)
	}
	if nxFlag {
		return commandResInt(1)
	} else {
		return commandResOk()
	}
}

func baseIncr(key, value string) *cmdResult {
	delta, err := strconv.Atoi(value)
	if err != nil {
		return commandResErrParseInt("value")
	}
	cmd := baseGet(key)
	if cmd.resType == resTypeFail {
		return cmd
	}
	var valueInt int
	if cmd.resType == resTypeNil {
		valueInt = 0
	} else {
		var err error
		valueInt, err = strconv.Atoi(cmd.resMsg)
		if err != nil {
			return commandResErrParseInt("value")
		}
	}
	valueInt = valueInt + delta
	baseSet(key, strconv.Itoa(valueInt), false, 0, false, false)
	return commandResInt(valueInt)
}

func baseDecr(key, value string) *cmdResult {
	delta, err := strconv.Atoi(value)
	if err != nil {
		return commandResErrParseInt("value")
	}
	cmd := baseGet(key)
	if cmd.resType == resTypeFail {
		return cmd
	}
	var valueInt int
	if cmd.resType == resTypeNil {
		valueInt = 0
	} else {
		var err error
		valueInt, err = strconv.Atoi(cmd.resMsg)
		if err != nil {
			return commandResErrParseInt("value")
		}
	}
	valueInt = valueInt - delta
	baseSet(key, strconv.Itoa(valueInt), false, 0, false, false)
	return commandResInt(valueInt)
}

func doAppend(opt ...string) *cmdResult {
	key := opt[0]
	value := opt[1]
	data, ex := getFromDb(key)
	var str string
	if ex {
		if data.dataType != dataNodeTypeString {
			return commandResErrType()
		}
		if str1, ok := data.dataPointer.(string); ok {
			str = str1
		} else {
			str = ""
		}
	} else {
		str = ""
	}
	str = str + value
	cmd := baseSet(key, str, false, 0, false, false)
	if cmd.resType != resTypeMsg {
		return cmd
	}
	return commandResInt(len(str))
}

func doBitCount(opt ...string) *cmdResult {
	key := opt[0]
	cmd := baseGet(key)
	var str string
	if cmd.resType == resTypeString {
		str = cmd.resMsg
	} else if cmd.resType == resTypeNil {
		return commandResInt(0)
	} else {
		return cmd
	}
	optLen := len(opt)
	len := len(str)
	var start, end int = 0, len - 1
	var err error
	if optLen == 3 {
		startStr := opt[1]
		start, err = strconv.Atoi(startStr)
		if err != nil {
			return commandResErrParseInt("value")
		}
		endStr := opt[2]
		end, err = strconv.Atoi(endStr)
		if err != nil {
			return commandResErrParseInt("value")
		}
	} else if optLen != 1 {
		return commandResErrSyntax()
	}
	if start < 0 {
		start = len + start
	}
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len + end
	}
	if end < 0 {
		end = 0
	}
	if end >= len {
		end = len - 1
	}
	if start > end {
		return commandResInt(0)
	}
	var count = 0
	for i := start; i <= end; i++ {
		b := uint8(str[i])
		for j := uint(0); j < 8; j++ {
			testByte := uint8(1 << j)
			if b&testByte > 0 {
				count++
			}
		}
	}
	return commandResInt(count)
}

func doBitOp(opt ...string) *cmdResult {
	opStr := strings.ToLower(opt[0])
	if opStr != "and" && opStr != "or" && opStr != "xor" && opStr != "not" {
		return commandResErrSyntax()
	}
	if opStr == "not" && len(opt) != 3 {
		return commandResErr("ERR BITOP NOT must be called with a single source key.")
	}
	setKey := opt[1]
	maxLen := 0
	maxLenPos := 0
	valuesList := make([]string, len(opt)-2)
	for i, key := range opt[2:] {
		cmd := baseGet(key)
		if cmd.resType == resTypeFail {
			return cmd
		}
		valuesList[i] = cmd.resMsg
		if maxLen < len(cmd.resMsg) {
			maxLen = len(cmd.resMsg)
			maxLenPos = i
		}
	}
	if opStr == "not" {
		res := make([]byte, len(valuesList[0]))
		for i := 0; i < len(valuesList[0]); i++ {
			b := valuesList[0][i]
			res[i] = 0xff ^ b
		}
		baseSet(setKey, string(res), false, 0, false, false)
		return commandResInt(len(res))
	}
	res := []byte(valuesList[maxLenPos])
	for i, str := range valuesList {
		if i == maxLenPos {
			continue
		}
		for j := 0; j < len(str); j++ {
			b := str[j]
			switch opStr {
			case "and":
				res[j] = res[j] & b
			case "or":
				res[j] = res[j] | b
			case "xor":
				res[j] = res[j] ^ b
			}
		}
	}
	baseSet(setKey, string(res), false, 0, false, false)
	return commandResInt(len(res))
}

func doBitPos(opt ...string) *cmdResult {
	key := opt[0]
	bitStr := opt[1]
	bit, err := strconv.Atoi(bitStr)
	if err != nil || (bit != 0 && bit != 1) {
		return commandResErrParseInt("bit")
	}
	cmd := baseGet(key)
	var str string
	if cmd.resType == resTypeString {
		str = cmd.resMsg
	} else if cmd.resType == resTypeNil {
		if bit == 0 {
			return commandResInt(0)
		} else {
			return commandResInt(-1)
		}
	} else {
		return cmd
	}
	optLen := len(opt)
	len := len(str)
	var start, end int = 0, len - 1
	var setEndFlag bool = false
	if optLen > 4 {
		return commandResErrSyntax()
	}
	if optLen > 2 {
		startStr := opt[2]
		start, err = strconv.Atoi(startStr)
		if err != nil {
			return commandResErrParseInt("value")
		}
		if optLen == 4 {
			endStr := opt[3]
			end, err = strconv.Atoi(endStr)
			if err != nil {
				return commandResErrParseInt("value")
			}
			setEndFlag = true
		}
	}
	if start < 0 {
		start = len + start
	}
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len + end
	}
	if end < 0 {
		end = 0
	}
	if end >= len {
		end = len - 1
	}
	if start > end {
		return commandResInt(-1)
	}
	var pos, findPos int = start * 8, -1
	for i := start; i <= end; i++ {
		b := str[i]
		if bit == 0 {
			b = b ^ 0xff
		}
		if b > 0 {
			findPos = pos
			for j := uint8(0x80); j > 0; j = j >> 1 {
				if j&b > 0 {
					break
				}
				findPos++
			}
			break
		}
		pos = pos + 8
	}
	if bit == 0 && findPos == -1 && start < len && setEndFlag == false {
		findPos = pos
	}
	return commandResInt(findPos)
}

func doDecr(opt ...string) *cmdResult {
	return baseDecr(opt[0], "1")
}

func doDecrBy(opt ...string) *cmdResult {
	return baseDecr(opt[0], opt[1])
}

func doGet(opt ...string) *cmdResult {
	key := opt[0]
	return baseGet(key)
}

func doGetBit(opt ...string) *cmdResult {
	key := opt[0]
	offsetStr := opt[1]
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return commandResErrParseInt("bit offset")
	}
	cmd := baseGet(key)
	var str string
	if cmd.resType == resTypeString {
		str = cmd.resMsg
	} else if cmd.resType == resTypeNil {
		str = ""
	} else {
		return cmd
	}
	pos := offset / 8
	len := len(str)
	if len <= pos {
		return commandResInt(0)
	}
	var testByte byte
	testByte = 1 << byte((7 - (offset % 8)))
	var res = 0
	if testByte&str[pos] > 0 {
		res = 1
	}
	return commandResInt(res)
}

func doGetRange(opt ...string) *cmdResult {
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
	cmd := baseGet(key)
	if cmd.resType == resTypeFail {
		return cmd
	}
	len := len(cmd.resMsg)
	if start < 0 {
		start = len + start
	}
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len + end
	}
	end = end + 1
	if end < start {
		return commandResString("")
	}
	if end > len {
		end = len
	}
	return commandResString(cmd.resMsg[start:end])
}

func doGetSet(opt ...string) *cmdResult {
	key := opt[0]
	value := opt[1]
	cmd := baseGet(key)
	if cmd.resType == resTypeFail {
		return cmd
	}
	baseSet(key, value, false, 0, false, false)
	return cmd
}

func doIncr(opt ...string) *cmdResult {
	return baseIncr(opt[0], "1")
}

func doIncrBy(opt ...string) *cmdResult {
	return baseIncr(opt[0], opt[1])
}

func doIncrByFloat(opt ...string) *cmdResult {
	key := opt[0]
	valueString := opt[1]
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return commandResErr("ERR value is not a valid float")
	}
	cmd := baseGet(key)
	if cmd.resType == resTypeFail {
		return cmd
	}
	var valueFloat float64
	if cmd.resType == resTypeNil {
		valueFloat = 0
	} else {
		var err error
		valueFloat, err = strconv.ParseFloat(cmd.resMsg, 64)
		if err != nil {
			return commandResErr("ERR value is not a valid float")
		}
	}
	valueFloat = valueFloat + value
	baseSet(key, strconv.FormatFloat(valueFloat, 'f', -1, 64), false, 0, false, false)
	return baseGet(key)
}

func doMGet(opt ...string) *cmdResult {
	resList := make([]*cmdResult, len(opt))
	for i, key := range opt {
		data, ex := getFromDb(key)
		if ex {
			if data.dataType != dataNodeTypeString {
				resList[i] = commandResErrType()
			}
			if str, ok := data.dataPointer.(string); ok {
				resList[i] = commandResString(str)
			} else {
				resList[i] = commandResString("")
			}
		} else {
			resList[i] = commandResNil()
		}
	}
	return commandResArray(resList)
}

func doMSet(opt ...string) *cmdResult {
	return baseMSet(false, opt...)
}

func doMSetNx(opt ...string) *cmdResult {
	return baseMSet(true, opt...)
}

func doPSetEx(opt ...string) *cmdResult {
	key := opt[0]
	value := opt[1]
	ttlStr := opt[2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	return baseSet(key, value, true, ttl, false, false)
}

func doSet(opt ...string) *cmdResult {
	key := opt[0]
	value := opt[1]
	nxFlag := false
	xxFlag := false
	ttlSetFlag := false
	ttlMs := 0
	for i := 2; i < len(opt); i++ {
		option := strings.ToLower(opt[i])
		switch option {
		case "nx":
			if nxFlag || xxFlag {
				return commandResErrSyntax()
			}
			nxFlag = true
		case "xx":
			if nxFlag || xxFlag {
				return commandResErrSyntax()
			}
			xxFlag = true
		case "ex":
			if ttlSetFlag {
				return commandResErrSyntax()
			}
			ttlSetFlag = true
			i++
			var err error
			ttlMs, err = strconv.Atoi(opt[i])
			if err != nil {
				return commandResErrParseInt("value")
			}
			ttlMs = ttlMs * 1000
		case "px":
			if ttlSetFlag {
				return commandResErrSyntax()
			}
			ttlSetFlag = true
			i++
			var err error
			ttlMs, err = strconv.Atoi(opt[i])
			if err != nil {
				return commandResErrParseInt("value")
			}
		default:
			return commandResErrSyntax()
		}
	}
	return baseSet(key, value, ttlSetFlag, ttlMs, nxFlag, xxFlag)
}

func doSetBit(opt ...string) *cmdResult {
	key := opt[0]
	posStr := opt[1]
	bitStr := opt[2]
	pos, err := strconv.Atoi(posStr)
	if err != nil || pos < 0 {
		return commandResErrParseInt("bit offset")
	}
	bit, err := strconv.Atoi(bitStr)
	if err != nil || (bit != 0 && bit != 1) {
		return commandResErrParseInt("bit")
	}
	cmd := baseGet(key)
	var str string
	if cmd.resType == resTypeString {
		str = cmd.resMsg
	} else if cmd.resType == resTypeNil {
		str = ""
	} else {
		return cmd
	}
	setLen := (pos / 8) + 1
	oldLen := len(str)
	capLen := setLen
	if capLen < oldLen {
		capLen = oldLen
	}
	cap := make([]byte, capLen)
	for i := 0; i < oldLen; i++ {
		cap[i] = str[i]
	}
	var testByte byte
	testByte = 1 << byte((7 - (pos % 8)))
	testPos := setLen - 1
	res := 0
	if testByte&cap[testPos] > 0 {
		res = 1
	}
	if bit == 0 {
		cap[testPos] = cap[testPos] & (0xFF ^ testByte)
	} else {
		cap[testPos] = cap[testPos] | testByte
	}
	baseSet(key, string(cap), false, 0, false, false)
	return commandResInt(res)
}

func doSetEx(opt ...string) *cmdResult {
	key := opt[0]
	value := opt[1]
	ttlStr := opt[2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		return commandResErrParseInt("value")
	}
	return baseSet(key, value, true, ttl*1000, false, false)
}

func doSetNx(opt ...string) *cmdResult {
	key := opt[0]
	value := opt[1]
	cmd := baseSet(key, value, false, 0, true, false)
	if cmd.resType == resTypeMsg {
		return commandResInt(1)
	} else if cmd.resType == resTypeNil {
		return commandResInt(0)
	}
	return cmd
}

func doSetRange(opt ...string) *cmdResult {
	key := opt[0]
	posStr := opt[1]
	value := opt[2]
	pos, err := strconv.Atoi(posStr)
	if err != nil || pos < 0 {
		return commandResErrParseInt("offset")
	}
	cmd := baseGet(key)
	var str string
	var ex bool
	if cmd.resType == resTypeString {
		str = cmd.resMsg
		ex = true
	} else if cmd.resType == resTypeNil {
		str = ""
		ex = false
	} else {
		return cmd
	}
	addLen := len(value)
	if addLen == 0 && ex == false {
		return commandResInt(0)
	}
	oldLen := len(str)
	capLen := addLen + pos
	if capLen < oldLen {
		capLen = oldLen
	}
	cap := make([]byte, capLen)
	for i := 0; i < oldLen; i++ {
		cap[i] = str[i]
	}
	for i := 0; i < addLen; i++ {
		cap[i+pos] = value[i]
	}
	baseSet(key, string(cap), false, 0, false, false)
	return commandResInt(len(string(cap)))
}

func doStrlen(opt ...string) *cmdResult {
	key := opt[0]
	cmd := baseGet(key)
	if cmd.resType == resTypeString {
		return commandResInt(len(cmd.resMsg))
	} else if cmd.resType == resTypeNil {
		return commandResInt(0)
	}
	return cmd
}
