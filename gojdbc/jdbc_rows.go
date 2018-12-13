package gojdbc

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
)

type JDBCRows struct {
	jrst C_JOBJECT
}

func (this *JDBCRows) Columns() []string {
	colText := pJDBCVM.CallStringMethod(this.jrst, arrMethod[mid_rst_column])
	return strings.Split(colText, "<*>")
}
func (this *JDBCRows) Close() error {
	if this.jrst != nil {
		pJDBCVM.CallVoidMethod(this.jrst, arrMethod[mid_rst_close])
		pJDBCVM.DeleteObject(this.jrst)
		this.jrst = nil
	}
	return nil
}

func (this *JDBCRows) Next(dest []driver.Value) error {
	jrow := pJDBCVM.CallObjectMethod(this.jrst, arrMethod[mid_rst_fetch])
	if jrow == nil {
		return this.getError()
	}
	defer pJDBCVM.DeleteObject(jrow)
	row, err := pJDBCVM.ExpandJObjectArray(jrow)
	if err != nil {
		return err
	}
	if len(row) == 0 {
		return io.EOF
	}
	if len(row) != len(dest) {
		return fmt.Errorf("scan column size can't matched")
	}
	for i := 0; i < len(dest); i++ {
		dest[i] = row[i]
	}
	return nil
}

func (this *JDBCRows) getError() error {
	if this.jrst == nil {
		return fmt.Errorf("JDBCRows:jrst is nil")
	}
	mErrInfo := pJDBCVM.CallStringMethod(this.jrst, arrMethod[mid_rst_geterror])
	if mErrInfo == "" {
		mErrInfo = "JDBCRows Unknown Error"
	}
	return fmt.Errorf(mErrInfo)
}
