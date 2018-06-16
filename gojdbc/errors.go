package gojdbc

import (
	"errors"
)

//错误定义
var (
	ErrQueryNoRow    = errors.New("no any row return") //查询行或查询列可能返回的错误结果
	ErrQueryConflict = errors.New("return isn't  recordset,use ExecXXXXXX replace")
	ErrExecConflict  = errors.New("return is recordset,use QueryXXXXXX replace")
	ErrQueryMaxParam = errors.New("unsupport call QueryMax with maxrows<=0")
)
