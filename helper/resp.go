package helper

import (
	"bytes"
	"github.com/hdt3213/rdb/model"
	"strconv"
)

const crlf = "\r\n"

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

func makeMultiBulkResp(args [][]byte) []byte {
	argLen := len(args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + crlf)
	for _, arg := range args {
		if arg == nil {
			buf.WriteString("$-1" + crlf)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + crlf + string(arg) + crlf)
		}
	}
	return buf.Bytes()
}

var setCmd = []byte("SET")

func stringToCmd(obj *model.StringObject) CmdLine {
	cmdLine := make([][]byte, 3)
	cmdLine[0] = setCmd
	cmdLine[1] = []byte(obj.Key)
	cmdLine[2] = obj.Value
	return cmdLine
}

var rPushAllCmd = []byte("RPUSH")

func listToCmd(obj *model.ListObject) CmdLine {
	cmdLine := make([][]byte, 2+obj.GetElemCount())
	cmdLine[0] = rPushAllCmd
	cmdLine[1] = []byte(obj.Key)
	for i, val := range obj.Values {
		cmdLine[2+i] = val
	}
	return cmdLine
}

var sAddCmd = []byte("SADD")

func setToCmd(obj *model.SetObject) CmdLine {
	cmdLine := make([][]byte, 2+obj.GetElemCount())
	cmdLine[0] = sAddCmd
	cmdLine[1] = []byte(obj.GetKey())
	for i, val := range obj.Members {
		cmdLine[2+i] = val
	}
	return cmdLine
}

var hMSetCmd = []byte("HMSET")

func hashToCmd(obj *model.HashObject) CmdLine {
	cmdLine := make([][]byte, 2+obj.GetElemCount()*2)
	cmdLine[0] = hMSetCmd
	cmdLine[1] = []byte(obj.GetKey())
	i := 0
	for field, val := range obj.Hash {
		cmdLine[2+i*2] = []byte(field)
		cmdLine[3+i*2] = val
		i++
	}
	return cmdLine
}

var zAddCmd = []byte("ZADD")

func zSetToCmd(obj *model.ZSetObject) CmdLine {
	cmdLine := make([][]byte, 2+obj.GetElemCount()*2)
	cmdLine[0] = zAddCmd
	cmdLine[1] = []byte(obj.GetKey())
	for i, e := range obj.Entries {
		value := strconv.FormatFloat(e.Score, 'f', -1, 64)
		cmdLine[2+i*2] = []byte(value)
		cmdLine[3+i*2] = []byte(e.Member)
	}
	return cmdLine
}

var pExpireAtBytes = []byte("PEXPIREAT")

// MakeExpireCmd generates command line to set expiration for the given key
func makeExpireCmd(obj model.RedisObject) CmdLine {
	expireAt := obj.GetExpiration()
	if expireAt == nil {
		return nil
	}
	args := make([][]byte, 3)
	args[0] = pExpireAtBytes
	args[1] = []byte(obj.GetKey())
	args[2] = []byte(strconv.FormatInt(expireAt.UnixNano()/1e6, 10))
	return args
}

// ObjectToCmd convert redis object to redis command line
func ObjectToCmd(obj model.RedisObject) []CmdLine {
	if obj == nil {
		return nil
	}
	cmdLines := make([]CmdLine, 0)
	switch obj.GetType() {
	case model.StringType:
		strObj := obj.(*model.StringObject)
		cmdLines = append(cmdLines, stringToCmd(strObj))
	case model.ListType:
		listObj := obj.(*model.ListObject)
		cmdLines = append(cmdLines, listToCmd(listObj))
	case model.HashType:
		hashObj := obj.(*model.HashObject)
		cmdLines = append(cmdLines, hashToCmd(hashObj))
	case model.SetType:
		setObj := obj.(*model.SetObject)
		cmdLines = append(cmdLines, setToCmd(setObj))
	case model.ZSetType:
		zsetObj := obj.(*model.ZSetObject)
		cmdLines = append(cmdLines, zSetToCmd(zsetObj))
	}
	if obj.GetExpiration() != nil {
		cmdLines = append(cmdLines, makeExpireCmd(obj))
	}
	return cmdLines
}

// CmdLinesToResp convert []CmdLine to RESP bytes
func CmdLinesToResp(cmds []CmdLine) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	for _, cmdLine := range cmds {
		resp := makeMultiBulkResp(cmdLine)
		buf.Write(resp)
	}
	return buf.Bytes()
}
