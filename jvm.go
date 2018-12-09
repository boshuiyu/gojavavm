package gojavavm

/*
#include "jvm_interface.h"
*/
import "C"
import (
	"database/sql/driver"
	"fmt"
	"os"
	"strconv"
	"time"
	"unsafe"
)

type JVM struct {
	pJavaVM unsafe.Pointer //java虚拟机对象
	jvmLib  string         //jvm.dll的路径
}

func PrintTime(info string) {
	fmt.Println(info + ":" + time.Now().Format("2006-01-02 15:04:05.000"))
}

//初始化jvm虚拟机,已经创建的不要重复创建
func (this *JVM) InitJVM(jarFile string) error {
	if this.pJavaVM != nil {
		return nil
	}
	this.jvmLib = this.QueryJVMLib()
	if this.jvmLib == "" {
		return fmt.Errorf("not config libarary path,create environment variable:GO_JVMLIB point to jvm.dll(windows) or libjvm.so(linux)")
	}
	fsinfo, err := os.Stat(this.jvmLib)
	if err != nil || fsinfo.IsDir() {
		return fmt.Errorf("can't found libarary,create environment variable:GO_JVMLIB point to jvm.dll(windows) or libjvm.so(linux)")
	}
	pszLib := C.CString(this.jvmLib)
	pszJar := C.CString(jarFile)
	defer C.free(unsafe.Pointer(pszLib))
	defer C.free(unsafe.Pointer(pszJar))
	var pszError *C.char = nil
	this.pJavaVM = C.CreateJavaVM(pszLib, pszJar, &pszError)
	err = SafePtrToError(pszError)
	if err != nil {
		this.UninitJVM()
		return err
	}
	return nil
}

func (this *JVM) IsLoaded() bool {
	return this.pJavaVM != nil
}

//释放jvm虚拟机,一般情况下不用
func (this *JVM) UninitJVM() {
	if this.pJavaVM != nil {
		C.DestroyJavaVM(this.pJavaVM)
		this.pJavaVM = nil
	}
}

//查找类
func (this *JVM) FindClass(clsName string) C_JCLASS {
	pszname := C.CString(clsName)
	defer C.free(unsafe.Pointer(pszname))
	return C.FindClass(this.pJavaVM, pszname)
}

//获得类的函数
func (this *JVM) GetMethodID(clazz C_JCLASS, name, sig string) C_JMETHODID {
	pszname := C.CString(name)
	pszsig := C.CString(sig)
	defer C.free(unsafe.Pointer(pszname))
	defer C.free(unsafe.Pointer(pszsig))
	res := C.GetMethodID(this.pJavaVM, clazz, pszname, pszsig)
	return C_JMETHODID(res)
}

//获得类的函数
func (this *JVM) GetStaticMethodID(clazz C_JCLASS, name, sig string) C_JMETHODID {
	pszname := C.CString(name)
	pszsig := C.CString(sig)
	defer C.free(unsafe.Pointer(pszname))
	defer C.free(unsafe.Pointer(pszsig))
	res := C.GetStaticMethodID(this.pJavaVM, clazz, pszname, pszsig)
	return C_JMETHODID(res)
}

//生成/删除对象
func (this *JVM) AllocObject(clazz C_JCLASS) C_JOBJECT {
	res := C.AllocGlobalObject(this.pJavaVM, clazz)
	return C_JOBJECT(res)
}
func (this *JVM) DeleteObject(obj C_JOBJECT) {

	if int(obj) != 0 {
		C.DeleteGlobalObject(this.pJavaVM, obj)
	}
}

//展开jobjectarray对象
func (this *JVM) ExpandJObjectArray(jobjArray C_JOBJECT) ([]interface{}, error) {
	var outSize C.int = 0
	pResult := C.ExpandJObjectArray(this.pJavaVM, jobjArray, &outSize)
	if int(outSize) < 0 {
		return nil, fmt.Errorf("not a jobjectarray")
	}
	//释放申请的内存
	defer func() {
		if pResult != nil {
			C.free(pResult)
		}
	}()
	if outSize == 0 {
		return nil, nil
	}
	//循环结果集转换类型
	arrResult := make([]interface{}, int(outSize))
	pAddr := uintptr(pResult)
	for i := 0; i < int(outSize); i++ {
		item := (*C.RESULT_ITEM)(unsafe.Pointer(pAddr))
		switch item.nType {
		case C.RETURN_TYPE_STRING:
			if item.pszData != nil {
				arrResult[i] = C.GoStringN((*C.char)(item.pszData), item.nLength)
				C.free(item.pszData)
			}
		case C.RETURN_TYPE_BYTES:
			if item.pszData != nil {
				arrResult[i] = C.GoBytes(item.pszData, item.nLength)
				C.free(item.pszData)
			}
		default:
			arrResult[i] = nil
		}
		pAddr += uintptr(C.RESULT_ITEM_SIZE)
	}
	return arrResult, nil
}

//调用java函数
func (this *JVM) CallVoidMethod(jobj C_JOBJECT, jmid C_JMETHODID, args ...interface{}) {
	this.wapperCallLongMethod(jobj, jmid, 'V', args...)
}
func (this *JVM) CallBoolMethod(jobj C_JOBJECT, jmid C_JMETHODID, args ...interface{}) bool {
	n := this.wapperCallLongMethod(jobj, jmid, 'Z', args...)
	return n == 1
}
func (this *JVM) CallIntMethod(jobj C_JOBJECT, jmid C_JMETHODID, args ...interface{}) int {
	n := this.wapperCallLongMethod(jobj, jmid, 'I', args...)
	return int(n)
}
func (this *JVM) CallLongMethod(jobj C_JOBJECT, jmid C_JMETHODID, args ...interface{}) int64 {
	n := this.wapperCallLongMethod(jobj, jmid, 'J', args...)
	return n
}
func (this *JVM) wapperCallLongMethod(jobj C_JOBJECT, jmid C_JMETHODID, jtype byte, args ...interface{}) int64 {
	params, paddr, isarray := this.createJVMParams(args...)
	defer this.deleteJVMParams(params)
	nRet := C.CallLongMethod(this.pJavaVM, jobj, jmid, C.char(jtype), paddr, C.int(len(params)), C.int(isarray))
	return int64(nRet)
}

//调用字符串返回值
func (this *JVM) CallStringMethod(jobj C_JOBJECT, jmid C_JMETHODID, args ...interface{}) string {
	params, paddr, isarray := this.createJVMParams(args...)
	defer this.deleteJVMParams(params)
	pszInfo := C.CallStringMethod(this.pJavaVM, jobj, jmid, paddr, C.int(len(params)), C.int(isarray))
	return SafePtrToString(pszInfo)
}

//调用object返回值
func (this *JVM) CallObjectMethod(jobj C_JOBJECT, jmid C_JMETHODID, args ...interface{}) C_JOBJECT {
	params, paddr, isarray := this.createJVMParams(args...)
	defer this.deleteJVMParams(params)
	return C.CallObjectMethod(this.pJavaVM, jobj, jmid, paddr, C.int(len(params)), C.int(isarray))
}

//调用字符串返回值
func (this *JVM) CallStaticStringMethod(clazz C_JCLASS, jmid C_JMETHODID, args ...interface{}) string {
	params, paddr, isarray := this.createJVMParams(args...)
	defer this.deleteJVMParams(params)
	pszInfo := C.CallStaticStringMethod(this.pJavaVM, clazz, jmid, paddr, C.int(len(params)), C.int(isarray))
	return SafePtrToString(pszInfo)
}

//调用object返回值
func (this *JVM) CallStaticObjectMethod(clazz C_JCLASS, jmid C_JMETHODID, args ...interface{}) C_JOBJECT {
	params, paddr, isarray := this.createJVMParams(args...)
	defer this.deleteJVMParams(params)
	return C.CallStaticObjectMethod(this.pJavaVM, clazz, jmid, paddr, C.int(len(params)), C.int(isarray))
}

//参数处理,内部调用
func (this *JVM) createJVMParams(args ...interface{}) ([]C.PARAM_ITEM, *C.PARAM_ITEM, int) {
	nCount := len(args)
	if nCount == 0 {
		return nil, nil, 0
	}
	isArray := 0
	//如果要以数组方式调用，只允许第唯一个参数是数组,并且都转成interface{}来处理
	var arrArgs []interface{} = args
	if v0 := args[0]; v0 != nil {
		if t, ok := v0.([]driver.Value); ok {
			isArray = 1
			nCount = len(t)
			arrArgs = make([]interface{}, nCount)
			for i, v := range t {
				arrArgs[i] = v
			}
		} else if t, ok := v0.([]interface{}); ok {
			isArray = 1
			nCount = len(t)
			arrArgs = t
		}
	}
	//再次检查长度
	if nCount == 0 {
		return nil, nil, isArray
	}
	paList := make([]C.PARAM_ITEM, nCount)
	for i, v := range arrArgs {
		if v == nil {
			paList[i].nType = C.PARAM_TYPE_NULL
		} else {
			paList[i].nType, paList[i].nLength, paList[i].nIntVal = this.getParamResult(v)
		}
	}
	return paList, &paList[0], isArray
}
func (this *JVM) getParamResult(v interface{}) (nType, nLength C.int, nValue C.INT64) {
	switch v.(type) {
	case []byte:
		nType = C.PARAM_TYPE_BYTES
		if bt, ok := v.([]byte); ok && len(bt) > 0 {
			nLength = C.int(len(bt))
			pBuffer := C.CBytes(bt)
			nValue = C.INT64(uintptr(pBuffer))
		}
	case int, int8, int16, int32, int64:
		nType = C.PARAM_TYPE_INT64
		i64, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		nValue = C.INT64(i64)
	case uint, uint8, uint16, uint32, uint64:
		nType = C.PARAM_TYPE_INT64
		u64, _ := strconv.ParseUint(fmt.Sprintf("%v", v), 10, 64)
		nValue = C.INT64(u64)
	default:
		stext := fmt.Sprint(v)
		nType = C.PARAM_TYPE_STRING
		nLength = C.int(len(stext))
		pBuffer := unsafe.Pointer(C.CString(stext))
		nValue = C.INT64(uintptr(pBuffer))
		//fmt.Printf("malloc string data:%p\r\n", pBuffer)
	}
	return
}

func (this *JVM) deleteJVMParams(params []C.PARAM_ITEM) {
	for _, item := range params {
		if item.nIntVal == 0 {
			continue
		}
		if item.nType == C.PARAM_TYPE_STRING {
			paddr := unsafe.Pointer(uintptr(item.nIntVal))
			C.free(paddr)
			//fmt.Printf("free string data:%p\r\n", paddr)
		} else if item.nType == C.PARAM_TYPE_BYTES {
			paddr := unsafe.Pointer(uintptr(item.nIntVal))
			C.free(paddr)
			//fmt.Printf("free byte data:%p\r\n", paddr)
		}
	}
}

/*
//转换为string,来源sql\convert.go
func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}
*/
/*
func (this *JVM) CreateJVMParams(args ...interface{}) (unsafe.Pointer, C.int) {
	nLen := len(args)
	if nLen <= 0 {
		return nil, C.int(0)
	}
	paList := C.newParamList(C.int(nLen))
	for index, v := range args {
		if v == nil {
			C.setParamItem(paList, C.int(index), C.PARAM_TYPE_NULL, 0, 0, nil)
		} else {
			switch v.(type) {
			case []driver.Value: //driver.Value数据库驱动中经常用到,特殊支持一下
				pszData, nLength := this.createValueArrayParam(v.([]driver.Value))
				C.setParamItem(paList, C.int(index), C.PARAM_TYPE_PARAMLIST, nLength, 0, pszData)
			case []interface{}:
				pszData, nLength := this.createInterfaceArrayParam(v.([]interface{}))
				C.setParamItem(paList, C.int(index), C.PARAM_TYPE_PARAMLIST, nLength, 0, pszData)
			default:
				nType, nLength, nValue, pszData := this.getParamResult(v)
				C.setParamItem(paList, C.int(index), nType, nLength, nValue, pszData)
			}
		}
	}
	return paList, C.int(nLen)
}
func (this *JVM) DeleteJVMParams(palist unsafe.Pointer, nsize C.int) {
	C.freeParamList(palist, nsize)
}

//参数辅助函数
func (this *JVM) getParamResult(v interface{}) (nType, nLength C.int, nValue C.INT64, pszData unsafe.Pointer) {
	nType, nLength, nValue, pszData = C.PARAM_TYPE_NULL, C.int(0), C.INT64(0), nil
	if v == nil {
		return
	}
	switch v.(type) {
	case float32:
	case []byte:
		nType = C.PARAM_TYPE_BYTES
		if bt, ok := v.([]byte); ok && len(bt) > 0 {
			nLength = C.int(len(bt))
			pszData = unsafe.Pointer(C.newByteBuffer(unsafe.Pointer(&bt[0]), nLength))
			fmt.Printf("malloc data:%p\r\n", pszData)
		}
	default:
		stext := fmt.Sprint(v)
		nType = C.PARAM_TYPE_STRING
		nLength = C.int(len(stext))
		pszData = unsafe.Pointer(C.CString(stext))
		fmt.Printf("malloc data:%p\r\n", pszData)
	}
	return
}
func (this *JVM) createInterfaceArrayParam(args []interface{}) (pszData unsafe.Pointer, nLength C.int) {
	//这个地方只需要获取到数据指针，不用保留数组对象。防止被GC释放，需要直接用C分配内存并管理
	nLength = C.int(len(args))
	palist := C.newParamList(nLength)
	for index, v := range args {
		nType, nLength, nValue, pszData := this.getParamResult(v)
		C.setParamItem(palist, C.int(index), nType, nLength, nValue, pszData)
	}
	return palist, nLength
}
func (this *JVM) createValueArrayParam(args []driver.Value) (pszData unsafe.Pointer, nLength C.int) {
	//这个地方只需要获取到数据指针，不用保留数组对象。防止被GC释放，需要直接用C分配内存并管理
	nLength = C.int(len(args))
	palist := C.newParamList(nLength)
	for index, v := range args {
		nType, nLength, nValue, pszData := this.getParamResult(v)
		C.setParamItem(palist, C.int(index), nType, nLength, nValue, pszData)
	}
	return palist, nLength
}
*/
