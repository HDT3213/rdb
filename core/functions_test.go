package core

import (
	"os"
	"testing"

	"github.com/hdt3213/rdb/model"
)



func TestParseFunctions(t *testing.T) {
	rdbFile, err := os.Open("../cases/function.rdb")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()

	dec := NewDecoder(rdbFile).WithSpecialOpCode()
	var functionLua string
	dec.Parse(func(object model.RedisObject) bool {
		if object.GetType() == model.FunctionsType {
			functionObj := object.(*model.FunctionsObject)
			functionLua = functionObj.FunctionsLua
			return false
		}
		return true
	})
	expect := "#!lua name=mylib\nredis.register_function('myfunc', function(keys, args) return 'hello' end)"
	if functionLua != expect {
		t.Error("function lua is not equals")
	}
}