package gojavavm

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

//试图自动化查找下jvm库文件
func (this *JVM) QueryJVMLib() string {
	//1.取环境变量
	jvmLib := queryJVMLibFromEnv()
	if jvmLib != "" {
		return jvmLib
	}
	//2.通过配置文件查找
	jvmLib = loadJVMFile()
	if jvmLib != "" {
		log.Println("通过dt文件找到libjvm.so", jvmLib)
		return jvmLib
	}
	//3.通过平台相关性查找
	jvmLib = loadJVMConfig()
	log.Println("通过平台相关性找到libjvm.so", jvmLib)
	return jvmLib
}

func loadJVMFile() string {
	var Utf8Bom = []byte{0xEF, 0xBB, 0xBF} //utf8文件头
	data, err := ioutil.ReadFile("./jvmlib_paths.dt")
	if err != nil {
		return ""
	}
	if bytes.HasPrefix(data, Utf8Bom) {
		data = data[len(Utf8Bom):]
	}
	for _, line := range strings.Split(string(data), "\n") {
		line := strings.Trim(line, "\r\n\t ")
		if line == "" {
			continue
		}
		if _, err := os.Stat(line); err == nil {
			return line
		}
	}
	return ""
}

func loadJVMConfig() string {

	arrLibJVMLocation := []string{
		//CentOS
		"/usr/share/ywjava/jre/lib/amd64/server/libjvm.so",
		//银河麒麟
		"/usr/lib/jvm/java-8-openjdk-arm64/jre/lib/aarch64/server/libjvm.so",
		//中标麒麟
		"/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.5/jre/lib/mips64/server/libjvm.so",
		//中科方德
		"/usr/lib/jvm/java-8-openjdk-amd64/jre/lib/amd64/server/libjvm.so",
	}

	//先根据制定路径查找
	for _, p := range arrLibJVMLocation {
		info, err := os.Stat(p)
		if err == nil && !info.IsDir() {
			return p
		}
	}

	//找不到的情况下，直接通过find命令查找
	return findFistLibJMVDirect()
}

//直接通过find命令查找	libjvm.so
func findFistLibJMVDirect() string {
	buf := bytes.NewBuffer(nil)
	cmdinfo := exec.Command("find", "/", "-name", "libjvm.so")
	cmdinfo.Stdout = buf
	err := cmdinfo.Run()
	if err != nil {
		return ""
	}

	arrLines := strings.Split(buf.String(), "\n")
	for _, line := range arrLines {
		info, err := os.Stat(line)
		if err == nil && !info.IsDir() {
			return strings.Trim(line, " \n")
		}
	}
	return ""
}
