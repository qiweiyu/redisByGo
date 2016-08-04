package core

type setsNodeData map[string]string

func createSetsNode(key string) *dataNode {
	var node = new(dataNode)
	node.key = key
	node.dataType = dataNodeTypeSet
	node.dataPointer = nil
	node.setTTL(0)
	return node
}

func baseSetsGet(key string) (setsNodeData, *cmdResult) {
	data, ex := getFromDb(key)
	if ex {
		if data.dataType != dataNodeTypeSet {
			return nil, commandResErrType()
		}
		if s, ok := data.dataPointer.(setsNodeData); ok {
			return s, nil
		}
		return nil, commandResErrType()
	} else {
		return nil, commandResNil()
	}
}

func baseSetsAdd(key string, value string) *cmdResult {
	var sets setsNodeData
	data, ex := getFromDb(key)
	if ex {
		if data.dataType != dataNodeTypeSet {
			return commandResErrType()
		}
		var ok bool
		if sets, ok = data.dataPointer.(setsNodeData); ok == false {
			return commandResErrType()
		}
	} else {
		sets = make(setsNodeData)
		data = createSetsNode(key)
	}
	if _, has := sets[value]; has {
		return commandResInt(0)
	}
	sets[value] = value
	data.dataPointer = interface{}(sets)
	setToDb(key, data)
	return commandResInt(1)
}

func baseSDiff(keys ...string) (setsNodeData, *cmdResult) {
	res, cmd := baseSUnion(keys[1:]...)
	if res == nil {
		return nil, cmd
	}
	var set0 setsNodeData
	set0, cmd = baseSetsGet(keys[0])
	if set0 == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return nil, cmd
		} else {
			return make(setsNodeData), nil
		}
	}
	diffRes := make(setsNodeData)
	for member := range set0 {
		if _, ex := res[member]; ex == false {
			diffRes[member] = member
		}
	}
	return diffRes, nil
}

func baseSInter(keys ...string) (setsNodeData, *cmdResult) {
	countList := make(map[string]int)
	for _, key := range keys {
		node, cmd := baseSetsGet(key)
		if node == nil {
			if cmd != nil && cmd.resType == resTypeFail {
				return nil, cmd
			}
			continue
		}
		for member := range node {
			if count, ex := countList[member]; ex {
				countList[member] = count + 1
			} else {
				countList[member] = 1
			}
		}
	}
	res := make(setsNodeData)
	for member, count := range countList {
		if count == len(keys) {
			res[member] = member
		}
	}
	return res, nil
}
func baseSUnion(keys ...string) (setsNodeData, *cmdResult) {
	res := make(setsNodeData)
	for _, key := range keys {
		node, cmd := baseSetsGet(key)
		if node == nil {
			if cmd != nil && cmd.resType == resTypeFail {
				return nil, cmd
			}
			continue
		}
		for member := range node {
			res[member] = member
		}
	}
	return res, nil
}

func doSAdd(opt ...string) *cmdResult {
	var count = 0
	key := opt[0]
	for _, value := range opt[1:] {
		cmd := baseSetsAdd(key, value)
		if cmd.resType != resTypeInt {
			return cmd
		}
		if cmd.resInt == 1 {
			count++
		}
	}
	return commandResInt(count)
}

func doSCard(opt ...string) *cmdResult {
	key := opt[0]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	return commandResInt(len(node))
}

func doSDiff(opt ...string) *cmdResult {
	res, cmd := baseSDiff(opt...)
	if res == nil {
		return cmd
	}
	resList := make([]*cmdResult, len(res))
	i := 0
	for member := range res {
		resList[i] = commandResString(member)
		i++
	}
	return commandResArray(resList)
}

func doSDiffStore(opt ...string) *cmdResult {
	key := opt[0]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
	}
	node, cmd = baseSDiff(opt[1:]...)
	if node == nil {
		return cmd
	}
	rmFromDb(key)
	for member := range node {
		baseSetsAdd(key, member)
	}
	return commandResInt(len(node))
}

func doSInter(opt ...string) *cmdResult {
	res, cmd := baseSInter(opt...)
	if res == nil {
		return cmd
	}
	resList := make([]*cmdResult, len(res))
	i := 0
	for member := range res {
		resList[i] = commandResString(member)
		i++
	}
	return commandResArray(resList)
}

func doSInterStore(opt ...string) *cmdResult {
	key := opt[0]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
	}
	node, cmd = baseSInter(opt[1:]...)
	if node == nil {
		return cmd
	}
	rmFromDb(key)
	for member := range node {
		baseSetsAdd(key, member)
	}
	return commandResInt(len(node))
}

func doSIsMember(opt ...string) *cmdResult {
	key := opt[0]
	member := opt[1]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	if _, ex := node[member]; ex {
		return commandResInt(1)
	}
	return commandResInt(0)
}

func doSMembers(opt ...string) *cmdResult {
	key := opt[0]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResEmptyArray()
	}
	resList := make([]*cmdResult, len(node))
	i := 0
	for member := range node {
		resList[i] = commandResString(member)
		i++
	}
	return commandResArray(resList)
}

func doSMove(opt ...string) *cmdResult {
	sk := opt[0]
	dk := opt[1]
	member := opt[2]
	nodeS, cmd := baseSetsGet(sk)
	if nodeS == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	nodeD, cmd := baseSetsGet(dk)
	if nodeD == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
	}
	if _, ex := nodeS[member]; ex == false {
		return commandResInt(0)
	}
	delete(nodeS, member)
	baseSetsAdd(dk, member)
	rmIfEmpty(sk)
	return commandResInt(1)
}

//func doSPop(opt ...string) *cmdResult {}

//func doSRandMember(opt ...string) *cmdResult {}

func doSRem(opt ...string) *cmdResult {
	key := opt[0]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
		return commandResInt(0)
	}
	count := 0
	for _, member := range opt[1:] {
		if _, ex := node[member]; ex {
			delete(node, member)
			count++
		}
	}
	rmIfEmpty(key)
	return commandResInt(count)
}

func doSUnion(opt ...string) *cmdResult {
	res, cmd := baseSUnion(opt...)
	if res == nil {
		return cmd
	}
	resList := make([]*cmdResult, len(res))
	i := 0
	for member := range res {
		resList[i] = commandResString(member)
		i++
	}
	return commandResArray(resList)
}

func doSUnionStore(opt ...string) *cmdResult {
	key := opt[0]
	node, cmd := baseSetsGet(key)
	if node == nil {
		if cmd != nil && cmd.resType == resTypeFail {
			return cmd
		}
	}
	node, cmd = baseSUnion(opt[1:]...)
	if node == nil {
		return cmd
	}
	rmFromDb(key)
	for member := range node {
		baseSetsAdd(key, member)
	}
	return commandResInt(len(node))
}
