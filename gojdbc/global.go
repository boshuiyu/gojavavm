package gojdbc

import (
	"database/sql"
	"errors"
	"fmt"
	"gojavavm"
	"sync"
)

//类型重定义
type C_JOBJECT = gojavavm.C_JOBJECT
type C_JMETHODID = gojavavm.C_JMETHODID
type C_JCLASS = gojavavm.C_JCLASS
type C_JVALUE = gojavavm.C_JVALUE

var SafePtrToError = gojavavm.SafePtrToError
var SafePtrToString = gojavavm.SafePtrToString

//定义一个全局的jvm虚拟机,整个jdbc只需要一个就可以了
var pJDBCVM = &gojavavm.JVM{}

//只执行一次的,初始化jvm虚拟机
var onceInitJavaVM sync.Once
var muxInitJavaVM sync.Mutex
var errInitJavaVM error = nil

//java jdbc类定义
const (
	cls_conn   = 0 + iota //connect对象
	cls_stmt              //statement对象
	cls_record            //resultset对象
	cls_max
)

//java jdbc函数定义
const (
	//connect相关
	mid_conn_init     = 0 + iota //连接对象的构造函数
	mid_conn_geterror            //获取错误
	mid_conn_open                //打开
	mid_conn_prepare             //准备一个statement
	mid_conn_begin               //开启事务
	mid_conn_commit              //提交事务
	mid_conn_rollback            //回滚事务
	mid_conn_close               //关闭
	//statement相关
	mid_stmt_geterror //获取错误
	mid_stmt_numinput //获取stmt参数个数
	mid_stmt_exec     //执行语句
	mid_stmt_query    //查询结果
	mid_stmt_close    //关闭statement
	//recordset相关
	mid_rst_geterror //获取错误相关
	mid_rst_column   //获取字段列表头
	mid_rst_next     //移动到下一条记录
	mid_rst_fetch    //获得一行
	mid_rst_close    //关闭结果集
	//结束
	mid_max
)

//定义java包中的函数
var (
	arrClass     [cls_max]C_JCLASS    //java的class对象
	arrMethod    [mid_max]C_JMETHODID //java的函数对象
	ErrUnsupport = errors.New("unsupport")
)

//初始化,注册驱动
func init() {
	sql.Register("gojdbc", &JDBCDriver{})
}

//检查虚拟机创建的状态
func checkJVMSatus() error {
	onceInitJavaVM.Do(func() {
		errInitJavaVM = loadJvmAndLibrary()
	})
	return errInitJavaVM
}

//全局参数，用
var gJarFilePath = "./mysql-bin.jar;./gojdbc.jar"
var gDriverClass = ""

//设置jar lib和驱动查询回调函数
func SetJDBCDriverJar(jarFile string, driverName string) {
	if jarFile != "" {
		gJarFilePath = jarFile
	}
	if driverName != "" {
		gDriverClass = driverName
	}
}

//初始化虚拟机并且获得需要的函数
func loadJvmAndLibrary() error {
	var err error
	if err = pJDBCVM.InitJVM(gJarFilePath); err != nil {
		return err
	}
	if arrClass[cls_conn] = pJDBCVM.FindClass("golang/java/GOConnect"); arrClass[cls_conn] == nil {
		return fmt.Errorf("can not found class:GOConnect")
	}
	if arrClass[cls_stmt] = pJDBCVM.FindClass("golang/java/GOStatement"); arrClass[cls_stmt] == nil {
		return fmt.Errorf("can not found class:GOStatement")
	}
	if arrClass[cls_record] = pJDBCVM.FindClass("golang/java/GORecordSet"); arrClass[cls_record] == nil {
		return fmt.Errorf("can not found class:GORecordSet")
	}
	cls_connect, cls_statement, cls_rows := arrClass[cls_conn], arrClass[cls_stmt], arrClass[cls_record]
	//保存connect函数
	if arrMethod[mid_conn_init] = pJDBCVM.GetMethodID(cls_connect, "<init>", "()V"); arrMethod[mid_conn_init] == nil {
		return fmt.Errorf("not found connect function <Init>")
	}
	if arrMethod[mid_conn_geterror] = pJDBCVM.GetMethodID(cls_connect, "GetError", "()Ljava/lang/String;"); arrMethod[mid_conn_geterror] == nil {
		return fmt.Errorf("not found connect function <GetError>")
	}
	if arrMethod[mid_conn_open] = pJDBCVM.GetMethodID(cls_connect, "Open", "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;J)Z"); arrMethod[mid_conn_open] == nil {
		return fmt.Errorf("not found connect function <Open>")
	}
	if arrMethod[mid_conn_close] = pJDBCVM.GetMethodID(cls_connect, "Close", "()V"); arrMethod[mid_conn_close] == nil {
		return fmt.Errorf("not found connect function <Close>")
	}
	if arrMethod[mid_conn_prepare] = pJDBCVM.GetMethodID(cls_connect, "Prepare", "(Ljava/lang/String;)Ljava/lang/Object;"); arrMethod[mid_conn_prepare] == nil {
		return fmt.Errorf("not found connect function <Prepare>")
	}
	if arrMethod[mid_conn_begin] = pJDBCVM.GetMethodID(cls_connect, "Begin", "()Z"); arrMethod[mid_conn_begin] == nil {
		return fmt.Errorf("not found connect function <Begin>")
	}
	if arrMethod[mid_conn_commit] = pJDBCVM.GetMethodID(cls_connect, "Commit", "()Z"); arrMethod[mid_conn_commit] == nil {
		return fmt.Errorf("not found connect function <Commit>")
	}
	if arrMethod[mid_conn_rollback] = pJDBCVM.GetMethodID(cls_connect, "Rollback", "()Z"); arrMethod[mid_conn_rollback] == nil {
		return fmt.Errorf("not found connect function <Rollback>")
	}
	//保存statement函数
	if arrMethod[mid_stmt_geterror] = pJDBCVM.GetMethodID(cls_statement, "GetError", "()Ljava/lang/String;"); arrMethod[mid_stmt_geterror] == nil {
		return fmt.Errorf("not found statment function <GetError>")
	}
	if arrMethod[mid_stmt_numinput] = pJDBCVM.GetMethodID(cls_statement, "NumInput", "()I"); arrMethod[mid_stmt_numinput] == nil {
		return fmt.Errorf("not found statment function <NumInput>")
	}
	if arrMethod[mid_stmt_exec] = pJDBCVM.GetMethodID(cls_statement, "Execute", "([Ljava/lang/Object;)I"); arrMethod[mid_stmt_exec] == nil {
		return fmt.Errorf("not found statment function <Execute>")
	}
	if arrMethod[mid_stmt_query] = pJDBCVM.GetMethodID(cls_statement, "Query", "([Ljava/lang/Object;)Ljava/lang/Object;"); arrMethod[mid_stmt_query] == nil {
		return fmt.Errorf("not found statment function <Query>")
	}
	if arrMethod[mid_stmt_close] = pJDBCVM.GetMethodID(cls_statement, "Close", "()V"); arrMethod[mid_stmt_close] == nil {
		return fmt.Errorf("not found statment function <Close>")
	}
	//保存recordset函数
	if arrMethod[mid_rst_geterror] = pJDBCVM.GetMethodID(cls_rows, "GetError", "()Ljava/lang/String;"); arrMethod[mid_rst_geterror] == nil {
		return fmt.Errorf("not found recordset function <GetError>")
	}
	if arrMethod[mid_rst_column] = pJDBCVM.GetMethodID(cls_rows, "Columns", "()Ljava/lang/String;"); arrMethod[mid_rst_column] == nil {
		return fmt.Errorf("not found recordset function <Columns>")
	}
	if arrMethod[mid_rst_next] = pJDBCVM.GetMethodID(cls_rows, "Next", "()I"); arrMethod[mid_rst_next] == nil {
		return fmt.Errorf("not found recordset function <Next>")
	}
	if arrMethod[mid_rst_fetch] = pJDBCVM.GetMethodID(cls_rows, "Fetch", "()Ljava/lang/Object;"); arrMethod[mid_rst_fetch] == nil {
		return fmt.Errorf("not found recordset function <Fetch>")
	}
	if arrMethod[mid_rst_close] = pJDBCVM.GetMethodID(cls_rows, "Close", "()V"); arrMethod[mid_rst_close] == nil {
		return fmt.Errorf("not found recordset function <Close>")
	}
	return nil
}
