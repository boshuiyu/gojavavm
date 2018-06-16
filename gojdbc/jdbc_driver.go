package gojdbc

import (
	"database/sql/driver"
	"fmt"
)

//驱动接口实现
type JDBCDriver struct {
	dsn string //连接字符串
}

//打开一个数据库连接
func (this *JDBCDriver) Open(dsn string) (driver.Conn, error) {
	err := checkJVMSatus()
	if err != nil {
		return nil, err
	}
	this.dsn = dsn
	jconn, err := this.newJavaConnect()
	if err != nil {
		return nil, err
	}
	conn := &JDBCConnect{jconn: jconn}
	err = conn.Open(this.dsn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (this *JDBCDriver) newJavaConnect() (C_JOBJECT, error) {
	jconn := pJDBCVM.AllocObject(arrClass[cls_conn])
	if jconn == nil {
		return nil, fmt.Errorf("NewConnect() failed")
	}
	return C_JOBJECT(jconn), nil
}
