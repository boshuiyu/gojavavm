package gojdbc

import (
	"database/sql/driver"
	"fmt"
)

type JDBCStatement struct {
	jstmt   C_JOBJECT
	sqltext string
}

func (this *JDBCStatement) Close() error {
	if this.jstmt != nil {
		pJDBCVM.CallVoidMethod(this.jstmt, arrMethod[mid_stmt_close])
		pJDBCVM.DeleteObject(this.jstmt)
		this.jstmt = nil
	}
	return nil
}
func (this *JDBCStatement) Exec(args []driver.Value) (driver.Result, error) {
	n := pJDBCVM.CallIntMethod(this.jstmt, arrMethod[mid_stmt_exec], args)
	if n == -1 {
		err := this.getError()
		return nil, err
	}
	return &JDBCResult{nAffectCount: n}, nil
}

func (this *JDBCStatement) Query(args []driver.Value) (driver.Rows, error) {
	jobj := pJDBCVM.CallObjectMethod(this.jstmt, arrMethod[mid_stmt_query], args)
	if jobj == nil {
		return nil, this.getError()
	}
	jdbcRecord := &JDBCRows{jrst: jobj}
	return jdbcRecord, nil
}
func (this *JDBCStatement) NumInput() int {
	n := pJDBCVM.CallIntMethod(this.jstmt, arrMethod[mid_stmt_numinput])
	return n
}
func (this *JDBCStatement) getError() error {
	if this.jstmt == nil {
		return fmt.Errorf("JDBCStatement:jstmt is nil")
	}
	mErrInfo := pJDBCVM.CallStringMethod(this.jstmt, arrMethod[mid_stmt_geterror])
	if mErrInfo == "" {
		mErrInfo = "JDBCStatement Unknown Error"
	}
	return fmt.Errorf(mErrInfo)
}

//exec result
type JDBCResult struct {
	nAffectCount int
}

func (this *JDBCResult) LastInsertId() (int64, error) {
	return 0, ErrUnsupport
}
func (this *JDBCResult) RowsAffected() (int64, error) {
	return int64(this.nAffectCount), nil
}
