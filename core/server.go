package core

func doFlushAll(_ ...string) *cmdResult {
	flushDb()
	return commandResOk()
}

func doFlushDb(_ ...string) *cmdResult {
	return commandResOk()
}
