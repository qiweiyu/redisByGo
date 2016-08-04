package core

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func Handle(conn net.Conn) {
	var res *cmdResult
	for {
		cmd, err := parse(conn)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}
		if cmd.name == "quit" {
			fmt.Println("Close From Client")
			conn.Close()
			return
		}
		if handler, ok := commandMap[cmd.name]; ok {
			paramsCount := len(cmd.params)
			if handler.argsCount >= 0 && paramsCount != handler.argsCount {
				res = commandResErrArguments(cmd.name)
			} else if handler.argsCount < 0 && paramsCount < -1*handler.argsCount {
				res = commandResErrArguments(cmd.name)
			} else {
				lock(cmd.name)
				res = handler.handler(cmd.params...)
				unlock(cmd.name)
			}
		} else {
			res = commandResNotFound(cmd.name)
		}
		conn.Write([]byte(res.String()))
	}
}

func parse(conn net.Conn) (*cmd, error) {
	buf := bufio.NewReader(conn)
	cmdLine, err := readLine(buf)
	if err != nil {
		return nil, newCmdError(readCmdError)
	}
	isOneLineType := strings.Index(cmdLine, "*") != 0
	var cmdInfo *cmd
	if isOneLineType {
		cmdInfo = parseOneLineCmd(cmdLine)
	} else {
		cmdInfo, err = parseMultiLinesCmd(cmdLine, buf)
		if err != nil {
			return nil, newCmdError(decodeCmdError)
		}
	}
	cmdInfo.name = strings.ToLower(cmdInfo.name)
	return cmdInfo, nil
}

func parseOneLineCmd(cmdLine string) *cmd {
	var cmdInfo = new(cmd)
	cmdLine = strings.TrimSpace(cmdLine)
	infoArr := strings.Split(cmdLine, " ")
	cmdInfo.name = infoArr[0]
	cmdInfo.params = infoArr[1:]
	return cmdInfo
}

func parseMultiLinesCmd(startLine string, buf *bufio.Reader) (*cmd, error) {
	argsCount, err := strconv.Atoi(startLine[1:])
	if err != nil || argsCount == 0 {
		return nil, newCmdError(decodeCmdError)
	}
	var list = make([]string, argsCount)
	for i := 0; i < argsCount; i++ {
		line, err := readLine(buf)
		if err != nil {
			return nil, newCmdError(readCmdError)
		}
		if strings.Index(line, "$") != 0 {
			return nil, newCmdError(decodeCmdError)
		}
		line, err = readLine(buf)
		if err != nil {
			return nil, newCmdError(readCmdError)
		}
		list[i] = line
	}
	var cmdInfo = new(cmd)
	cmdInfo.name = list[0]
	cmdInfo.params = list[1:]
	return cmdInfo, nil
}

func readLine(buf *bufio.Reader) (string, error) {
	var str []byte
	for {
		strTmp, isPrefix, err := buf.ReadLine()
		str = append(str, strTmp...)
		if err != nil {
			return string(str), err
		}
		if isPrefix == false {
			break
		}
	}
	var s string
	s = string(str)
	return s, nil
}
