package gojavavm

import (
	"os"
	"os/exec"
	"path/filepath"
)

//试图自动化查找下jvm库文件
func (this *JVM) QueryJVMLib() string {
	//1.取环境变量
	jvmLib := queryJVMLibFromEnv()
	if jvmLib != "" {
		return jvmLib
	}
	//2.通过exe可执行路径查找
	file, err := exec.LookPath("java")
	if err != nil {
		return ""
	}
	javaFile, err := filepath.EvalSymlinks(file)
	if err != nil {
		javaFile = file
	}
	//指向当前目录下的client/jvm.dll
	jvmLib = filepath.Join(filepath.Dir(javaFile), "client\\jvm.dll")
	if _, err = os.Stat(jvmLib); err == nil {
		return jvmLib
	}
	return ""
}
