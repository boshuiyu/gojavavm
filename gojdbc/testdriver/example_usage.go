package main

/*
说明:
使用该jar驱动的环境要求:
1.将gojdbc.jar放到跟exe文件同目录下，
2.需要安装java环境，或将java中的jre目录拷贝到同exe文件目录中
目前支持的驱动类型
com.microsoft.sqlserver.jdbc.SQLServerDriver url like->jdbc:sqlserver://127.0.0.1:1433;databaseName=dbname
com.mysql.jdbc.Driver url like->jdbc:mysql://127.0.0.1:3306/dbname
oracle.jdbc.driver.OracleDriver url like->jdbc:oracle:thin:@192.168.2.127:1521:Orcl
具体使用参考下面的例子
*/
import (
	"database/sql"
	"fmt"
	_ "gojavavm/gojdbc"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var conn *sql.DB = nil
var nrTestIndex = 0 //总的测试循环次数
var tableName = "mytable"

func PrintTime(info string) {
	fmt.Println(info + ":" + time.Now().Format("2006-01-02 15:04:05.000"))
}
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("程序开始...")
	runtime.GC()
	defer log.Println("程序结束")
	err := open("gojdbc", "jdbc:mysql://127.0.0.1:3306/mydbtest?pp=2&pass=123456&user=root&allowMultiQueries=true&querytimeout=5&characterEncoding=utf8")
	if err != nil {
		log.Println("打开数据库错误:", err)
		close()
		return
	}
	defer close()
	createTable()
	for {
		err = testInsert(10, 10000)
	}
	checkErr(err)
	for {
		testQuery(rand.Intn(100) + 1)
	}

	log.Println("测试完成.....")
	time.Sleep(1 * time.Hour)
}

//测试打开关闭链接
func open(driverName, sourceData string) error {
	var err error = nil
	conn, err = sql.Open(driverName, sourceData)
	if err != nil {
		return err
	}
	conn.SetMaxIdleConns(10)
	//_, err = conn.Exec("delete from mytable where id=-1")
	return err
}

func close() {
	if conn != nil {
		conn.Close()
		conn = nil
	}
}

//创建表
func createTable() {
	sqltext := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
id  int(11) NOT NULL AUTO_INCREMENT ,
val_int  int(11) NULL DEFAULT 0,
val_bigint  bigint(20) NULL DEFAULT 0 ,
val_string  varchar(2000) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL ,
val_text  text CHARACTER SET utf8 COLLATE utf8_general_ci NULL ,
val_blob  blob NULL ,
val_datetime  datetime NULL DEFAULT NULL ,
PRIMARY KEY (id)
)
ENGINE=MyISAM
DEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci
AUTO_INCREMENT=3364
CHECKSUM=0
ROW_FORMAT=DYNAMIC
DELAY_KEY_WRITE=0
;`
	_, err := conn.Exec(sqltext)
	checkErr(err)
	//var value int
	//err = conn.QueryRow("select sleep(6)").Scan(&value)
	//fmt.Println("SLEEP的结果:", err, value)
	_, err = conn.Exec("truncate table " + tableName)
	checkErr(err)
	//_, err = conn.Exec("load_local_file ?", "c:\\abc.sql")
	//checkErr(err)
}

var sqlTextInsert = "insert into " + tableName + "(val_int,val_bigint,val_string,val_text,val_blob,val_datetime) values(?,?,?,?,?,?)"
var mFieldString = "_表名称被指定为db_name.tbl_name，以便在特定的数据库中创建表。不论是否有当前数据库，都可以通过这种方式创建表。如果您使用加引号的识别名，则应对数据库和表名称分别加引号。例如，`mydb`.`mytbl`是合法的，但是`mydb.mytbl`不合法。"

func testInsert(nThread int, nCount int) error {
	nrTestIndex++
	PrintTime("测试插入数据开始,第:" + strconv.Itoa(nrTestIndex) + "次...")
	defer func() {
		PrintTime("测试插入数据结束")
	}()
	plCount := new(int32)
	atomic.StoreInt32(plCount, int32(nCount))
	var wg sync.WaitGroup
	//定义处理函数
	var doThread = func(index int) {
		defer wg.Done()
		stmt, err := conn.Prepare(sqlTextInsert)
		checkErr(err)
		defer stmt.Close()
		var mIndex string
		for {
			i := atomic.AddInt32(plCount, -1)
			if i < 0 {
				break
			}
			mIndex = strconv.Itoa(int(i))
			valText := mIndex + mFieldString
			_, err := stmt.Exec(i, i+1000000000,
				valText, valText, []byte(valText), time.Now().Format("2006-01-02 15:04:05.000"))
			if err != nil {
				panic(err)
			}
		}
	}
	for i := 0; i < nThread; i++ {
		wg.Add(1)
		go doThread(i)
	}
	wg.Wait()
	return nil
}

func testQuery(nRowCount int) error {
	nrTestIndex++
	PrintTime("测试查询数据开始,第:" + strconv.Itoa(nrTestIndex) + "次...")
	defer func() {
		PrintTime("测试查询数据结束")
	}()
	rows, err := conn.Query("select * from mytable limit " + strconv.Itoa(nRowCount))
	checkErr(err)
	defer rows.Close()
	columns, err := rows.Columns()
	checkErr(err)
	colSize := len(columns)
	if colSize <= 0 {
		panic("字段列表为空")
	}
	var nrRecordCount = 0
	//填充数据
	scanArgs := make([]interface{}, colSize)
	rowIndex := 0
	for rows.Next() {
		values := make([]interface{}, colSize)
		for j := range values {
			scanArgs[j] = &values[j]
		}
		err = rows.Scan(scanArgs...)
		checkErr(err)
		for i := 0; i < colSize; i++ {
			//fmt.Println("row", rowIndex, ":", columns[i], "=", values[i])
		}
		rowIndex++
		nrRecordCount++
	}
	log.Println("记录集:", nrRecordCount, ",字段列表:", columns)
	return nil
}

//验证错误
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//测试连接和关闭
/*
func TestGOExampleGOJDBC() {

	for i := 0; i < 2; i++ {
		go doWorkTest(i * 100)
	}
	time.Sleep(10 * time.Hour)
}

func doWorkTest(index int) {
	var err error
	time.Sleep(time.Duration(index) * time.Millisecond)
	//构造连接池对象
	dbpool, err := gojdbc.NewJDBCPool()
	if err != nil {
		log.Println("获取数据库连接池失败", err)
		return
	}
	//设置连接驱动
	err = dbpool.SetDriver("com.microsoft.sqlserver.jdbc.SQLServerDriver")
	if err != nil {
		log.Println("设置驱动失败:", err)
		return
	}
	//设置连接池的最小，最大数量，不设置默最大最小只有一个连接，即单连接,需要在OPEN之前调用
	dbpool.SetLimit(1, 1)
	//设置执行查询语句的超时时间
	err = dbpool.SetTimeout(10)
	if err != nil {
		log.Println("设置执行超时时间错误:", err)
		return
	}
	//连接数据库
	err = dbpool.Open("jdbc:sqlserver://127.0.0.1:1433;databaseName=ybstest", "sa", "yw123456")
	if err != nil {
		log.Println("连接数据库失败:", err)
	}

	for {
		//查询数据
		_, _, err = dbpool.Query("select * from test where id>?", 0)
		if err != nil {
			log.Println("查询数据失败:", err)
			os.Exit(-1)
		}
		//执行命令
		_, err = dbpool.Exec("delete from test where id=? and ip=?", -1, "192.168.0.1")
		if err != nil {
			log.Println("查询数据失败:", err)
			os.Exit(-1)
		}
		time.Sleep(10 * time.Millisecond)
		log.Println("ROUTE:", index, "SUCCESS")
		os.Exit(-1)
	}
}
*/
