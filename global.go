package gojavavm

/*
#include "jvm_interface.h"
*/
import "C"
import (
	"errors"
	"os"
	"unsafe"
)

//类型重定义
type C_JOBJECT = C.jobject
type C_JMETHODID = C.jmethodID
type C_JCLASS = C.jclass
type C_JVALUE = C.jvalue

//释放由c++ malloc分配的内存,并返回字符串
func SafePtrToError(pszError *C.char) error {
	if pszError == nil {
		return nil
	}
	errString := C.GoString(pszError)
	C.FreeBuffer(unsafe.Pointer(pszError))
	return errors.New(errString)
}

func SafePtrToString(pszInfo *C.char) string {
	if pszInfo == nil {
		return ""
	}
	mInfo := C.GoString(pszInfo)
	C.FreeBuffer(unsafe.Pointer(pszInfo))
	return mInfo
}
func queryJVMLibFromEnv() string {
	jvmlib := os.Getenv("YWMC_JVMLIBA")
	if jvmlib != "" {
		return jvmlib
	}
	jvmlib = os.Getenv("GO_JVMLIB")
	if jvmlib != "" {
		return jvmlib
	}
	return ""
}
