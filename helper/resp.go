package helper

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/hdt3213/rdb/model"
)

const crlf = "\r\n"

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

// lexOrder traversal map in lex order to create
type lexOrder struct{}

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
var hPExpireAtCmd = []byte("HPEXPIREAT") // redis 7.4.0+
var hPersistCmd = []byte("HPERSIST")     // redis 7.4.0+

func hashToCmd(obj *model.HashObject, useLexOrder bool) []CmdLine {
	cmdLine := make([][]byte, 2+obj.GetElemCount()*2)
	cmdLine[0] = hMSetCmd
	cmdLine[1] = []byte(obj.GetKey())
	i := 0
	if useLexOrder {
		entries := make([][2][]byte, 0, 2*len(obj.Hash))
		for field, val := range obj.Hash {
			entries = append(entries, [2][]byte{
				[]byte(field), val,
			})
		}
		sort.Slice(entries, func(i, j int) bool {
			return string(entries[i][0]) < string(entries[j][0])
		})
		for _, entry := range entries {
			cmdLine[2+i*2] = entry[0]
			cmdLine[3+i*2] = entry[1]
			i++
		}
	} else {
		for field, val := range obj.Hash {
			cmdLine[2+i*2] = []byte(field)
			cmdLine[3+i*2] = val
			i++
		}
	}

	cmds := []CmdLine{cmdLine}
	if len(obj.FieldExpirations) == len(obj.Hash) {
		for field, expire := range obj.FieldExpirations {
			if expire == 0 {
				hpexp := make([][]byte, 5)
				// HPEXPIRE key seconds FIELDS num FIELD...
				hpexp[0] = hPersistCmd
				hpexp[1] = []byte(obj.Key)
				hpexp[2] = []byte("FIELDS")
				hpexp[3] = []byte("1")
				hpexp[4] = []byte(field)
				cmds = append(cmds, hpexp)
			} else {
				hpexp := make([][]byte, 6)
				// HPEXPIRE key seconds FIELDS num FIELD...
				hpexp[0] = hPExpireAtCmd
				hpexp[1] = []byte(obj.Key)
				hpexp[2] = []byte(fmt.Sprintf("%d", expire))
				hpexp[3] = []byte("FIELDS")
				hpexp[4] = []byte("1")
				hpexp[5] = []byte(field)
				cmds = append(cmds, hpexp)
			}
		}
	}

	return cmds
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

var (
	xaddCmd = []byte("XADD")
)

func formatStreamID(streamID *model.StreamId) string {
	ms := strconv.FormatUint(streamID.Ms, 10)
	seq := strconv.FormatUint(streamID.Sequence, 10)
	return ms + "-" + seq
}

func streamToCmd(stream *model.StreamObject) []CmdLine {
	commands := make([]CmdLine, 0)

	for _, entry := range stream.Entries {
		for _, message := range entry.Msgs {
			args := make(CmdLine, 0, 3+len(message.Fields)*2)
			args = append(
				args,
				xaddCmd,
				[]byte(stream.GetKey()),
				[]byte(formatStreamID(message.Id)),
			)

			for key, value := range message.Fields {
				args = append(args, []byte(key), []byte(value))
			}

			commands = append(commands, args)

			// TODO: groups, consumers, pending
		}
	}

	return commands
}

// ObjectToCmd convert redis object to redis command line
func ObjectToCmd(obj model.RedisObject, opts ...interface{}) []CmdLine {
	if obj == nil {
		return nil
	}
	useLexOrder := false
	for _, o := range opts {
		switch o.(type) {
		case lexOrder:
			useLexOrder = true
		}
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
		cmdLines = append(cmdLines, hashToCmd(hashObj, useLexOrder)...)
	case model.SetType:
		setObj := obj.(*model.SetObject)
		cmdLines = append(cmdLines, setToCmd(setObj))
	case model.ZSetType:
		zsetObj := obj.(*model.ZSetObject)
		cmdLines = append(cmdLines, zSetToCmd(zsetObj))
	case model.StreamType:
		streamObj := obj.(*model.StreamObject)
		cmdLines = append(cmdLines, streamToCmd(streamObj)...)
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

// WriteObjectToResp convert object to resp and write
func WriteObjectToResp(w io.Writer, obj model.RedisObject) error {
	var err error
	cmdLines := ObjectToCmd(obj)

	for _, cmdLine := range cmdLines {
		argLen := len(cmdLine)
		respBytes := make([][]byte, 0)
		respBytes = append(respBytes, []byte("*"+strconv.Itoa(argLen)+crlf))

		for _, arg := range cmdLine {
			if arg == nil {
				respBytes = append(respBytes, []byte("$-1"+crlf))
				continue
			}
			respBytes = append(respBytes, []byte("$"+strconv.Itoa(len(arg))+crlf))
			// the arg may be a large key or value
			respBytes = append(respBytes, arg)
			respBytes = append(respBytes, []byte(crlf))
		}

		for _, bytes := range respBytes {
			if _, err = w.Write(bytes); err != nil {
				return err
			}
		}
	}
	return nil
}
