package core

import (
	"bytes"
	"strconv"
)

const (
	resTypeMsg    = 1
	resTypeFail   = 2
	resTypeNil    = 3
	resTypeString = 4
	resTypeInt    = 5
	resTypeArray  = 6
)

type cmdResult struct {
	resType  int
	resMsg   string
	resInt   int
	resArray []*cmdResult
}

func (this *cmdResult) String() string {
	var str string
	switch this.resType {
	case resTypeMsg:
		buf := bytes.Buffer{}
		buf.WriteString("+")
		buf.WriteString(this.resMsg)
		buf.WriteString("\r\n")
		str = buf.String()
	case resTypeFail:
		buf := bytes.Buffer{}
		buf.WriteString("-")
		buf.WriteString(this.resMsg)
		buf.WriteString("\r\n")
		str = buf.String()
	case resTypeNil:
		str = "$-1\r\n"
	case resTypeString:
		buf := bytes.Buffer{}
		buf.WriteString("$")
		buf.WriteString(strconv.Itoa(len(this.resMsg)))
		buf.WriteString("\r\n")
		buf.WriteString(this.resMsg)
		buf.WriteString("\r\n")
		str = buf.String()
	case resTypeInt:
		buf := bytes.Buffer{}
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(this.resInt))
		buf.WriteString("\r\n")
		str = buf.String()
	case resTypeArray:
		buf := bytes.Buffer{}
		buf.WriteString("*")
		buf.WriteString(strconv.Itoa(len(this.resArray)))
		buf.WriteString("\r\n")
		for _, data := range this.resArray {
			buf.WriteString(data.String())
		}
		str = buf.String()
	}
	return str
}

func commandResOk() *cmdResult {
	res := new(cmdResult)
	res.resType = resTypeMsg
	res.resMsg = "OK"
	return res
}

func commandResNil() *cmdResult {
	res := new(cmdResult)
	res.resType = resTypeNil
	res.resMsg = ""
	res.resInt = 0
	res.resArray = nil
	return res
}

func commandResString(data string) *cmdResult {
	res := new(cmdResult)
	res.resType = resTypeString
	res.resMsg = data
	return res
}

func commandResInt(data int) *cmdResult {
	res := new(cmdResult)
	res.resType = resTypeInt
	res.resInt = data
	return res
}

func commandResArray(data []*cmdResult) *cmdResult {
	res := new(cmdResult)
	res.resType = resTypeArray
	res.resArray = data
	return res
}

func commandResEmptyArray() *cmdResult {
	return commandResArray(nil)
}

func commandResNotFound(name string) *cmdResult {
	return commandResErr("ERR unknown command '" + name + "'")
}

func commandResErrArguments(name string) *cmdResult {
	return commandResErr("ERR wrong number of arguments for '" + name + "' command")
}

func commandResErrSyntax() *cmdResult {
	return commandResErr("ERR syntax error")
}

func commandResErrParseInt(name string) *cmdResult {
	return commandResErr("ERR " + name + " is not an integer or out of range")
}

func commandResErrType() *cmdResult {
	return commandResErr("WRONGTYPE Operation against a key holding the wrong kind of value")
}

func commandResErr(msg string) *cmdResult {
	res := new(cmdResult)
	res.resType = resTypeFail
	res.resMsg = msg
	return res
}
