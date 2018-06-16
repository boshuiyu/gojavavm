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
	file, err := exec.LookPath("java.exe")
	if err != nil {
		return ""
	}
	javaFile, err := filepath.EvalSymlinks(file)
	if err != nil {
		javaFile = file
	}
	//指向当前目录下的client/jvm.dll
	javaBinDir := filepath.Dir(javaFile)
	//jre目录是这样对应的
	jvmLib = filepath.Join(javaBinDir, "./client/jvm.dll")
	if _, err = os.Stat(jvmLib); err == nil {
		return jvmLib
	}
	//jdk/bin目录是这样对应的
	jvmLib = filepath.Join(javaBinDir, "../jre/bin/client/jvm.dll")
	if _, err = os.Stat(jvmLib); err == nil {
		return jvmLib
	}
	return ""
}
