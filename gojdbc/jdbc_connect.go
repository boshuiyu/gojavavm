package gojdbc

import "C"
import (
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"strings"
	//"runtime"
	"strconv"
	"time"
)

func PrintTime(info string) {
	fmt.Println(info + ":" + time.Now().Format("2006-01-02 15:04:05.000"))
}

//url参数中驱动支持处理的参数
var supportURLArgs = []string{"user", "pass", "base64user", "base64pass", "querytimeout"}

//数据库链接实现
type JDBCConnect struct {
	jconn C_JOBJECT
}

func (this *JDBCConnect) Open(dsn string) error {
	var user, pass, timeout string
	var nTimeout = 0
	var ok bool
	mapParam, newDsn := this.fetchURLParam(dsn)
	//先提取用户名和密码,用户名密码可用base64编码传递
	timeout = mapParam["querytimeout"]
	if user, ok = mapParam["user"]; !ok {
		if user = mapParam["base64user"]; user != "" {
			if bt, e1 := base64.StdEncoding.DecodeString(user); e1 == nil {
				user = string(bt)
			} else {
				return fmt.Errorf("can't unbase64 param base64user")
			}
		}
	}
	if pass, ok = mapParam["pass"]; !ok {
		if pass = mapParam["base64pass"]; pass != "" {
			if bt, e1 := base64.StdEncoding.DecodeString(pass); e1 == nil {
				pass = string(bt)
			} else {
				return fmt.Errorf("can't unbase64 param base64pass")
			}
		}
	}
	if timeout != "" {
		if n, err := strconv.Atoi(timeout); err == nil && n > 0 {
			nTimeout = n
		}
	}
	ok = pJDBCVM.CallBoolMethod(this.jconn, arrMethod[mid_conn_open], newDsn, user, pass, int(nTimeout))
	if !ok {
		return this.getError()
	}
	return nil
}

func (this *JDBCConnect) Close() error {
	if this.jconn != nil {
		pJDBCVM.CallVoidMethod(this.jconn, arrMethod[mid_conn_close])
		pJDBCVM.DeleteObject(this.jconn)
		this.jconn = nil
	}
	return nil
}

func (this *JDBCConnect) Prepare(query string) (driver.Stmt, error) {
	jobj := pJDBCVM.CallObjectMethod(this.jconn, arrMethod[mid_conn_prepare], query)
	if jobj == nil {
		err := this.getError()
		return nil, err
	}
	jdbcStmt := &JDBCStatement{jstmt: jobj, sqltext: query}
	return jdbcStmt, nil
}

//事务相关
func (this *JDBCConnect) Begin() (driver.Tx, error) {
	ok := pJDBCVM.CallBoolMethod(this.jconn, arrMethod[mid_conn_begin])
	if !ok {
		err := this.getError()
		return nil, err
	}
	jdbcTx := &JDBCTx{pConn: this}
	return jdbcTx, nil
}
func (this *JDBCConnect) Commit() error {
	ok := pJDBCVM.CallBoolMethod(this.jconn, arrMethod[mid_conn_commit])
	if !ok {
		return this.getError()
	}
	return nil
}
func (this *JDBCConnect) Rollback() error {
	ok := pJDBCVM.CallBoolMethod(this.jconn, arrMethod[mid_conn_rollback])
	if !ok {
		return this.getError()
	}
	return nil
}

//获取jdbcconnect的错误
func (this *JDBCConnect) getError() error {
	if this.jconn == nil {
		return fmt.Errorf("JDBCConnect:jconn is nil")
	}
	mErrInfo := pJDBCVM.CallStringMethod(this.jconn, arrMethod[mid_conn_geterror])
	if mErrInfo == "" {
		mErrInfo = "JDBCConnect Unknown Error"
	}
	return fmt.Errorf(mErrInfo)
}

//提取出来驱动需要识别处理的参数,并返回新的连接字符串
func (this *JDBCConnect) fetchURLParam(orgdsn string) (mapResult map[string]string, newdsn string) {
	mapResult = make(map[string]string)
	newdsn = orgdsn
	//找到?后面的参数,用&分隔
	npos := strings.Index(orgdsn, "?")
	if npos == -1 {
		return
	}
	partAddr := orgdsn[:npos]
	partArgs := orgdsn[npos+1:]
	oldLines := strings.Split(partArgs, "&")
	newLines := make([]string, 0, len(oldLines))
	for _, line := range oldLines {
		var isFound = false
		for _, findKey := range supportURLArgs {
			if strings.HasPrefix(line, findKey+"=") {
				//这个是需要查找的值,放到map表里
				mapResult[findKey] = line[len(findKey)+1:]
				isFound = true
				break
			}
		}
		if !isFound { //不认识的，放到新的数组里
			newLines = append(newLines, line)
		}
	}
	if len(mapResult) == 0 {
		return
	}
	//重新拼接参数形成新的dsn
	newdsn = partAddr
	if len(newLines) > 0 {
		newdsn += "?" + strings.Join(newLines, "&")
	}
	return
}

//事务的简易实现
type JDBCTx struct {
	pConn *JDBCConnect
}

func (this *JDBCTx) Commit() error {
	return this.pConn.Commit()
}
func (this *JDBCTx) Rollback() error {
	return this.pConn.Rollback()
}
