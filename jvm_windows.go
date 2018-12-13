package gojavavm

import (
	"os"
	"sort"
	"syscall"
	"unsafe"
)

//试图自动化查找下jvm库文件
func (this *JVM) QueryJVMLib() string {
	//1.取环境变量
	jvmLib := queryJVMLibFromEnv()
	if jvmLib != "" {
		return jvmLib
	}
	//2.通过注册表查找
	var mPath = "SOFTWARE\\JavaSoft\\Java Runtime Environment"
	var hRoot syscall.Handle = 0
	err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, syscall.StringToUTF16Ptr(mPath), 0, syscall.KEY_READ, &hRoot)
	if err != nil {
		return ""
	}
	defer syscall.RegCloseKey(hRoot)
	var arrSubKey = []string{}
	var index uint32 = 0
	for {
		nameLen := uint32(1024)
		nameBuf := make([]uint16, nameLen)
		err = syscall.RegEnumKeyEx(hRoot, index, &nameBuf[0], &nameLen, nil, nil, nil, nil)
		if err != nil {
			break
		}
		index++
		arrSubKey = append(arrSubKey, mPath+"\\"+syscall.UTF16ToString(nameBuf))
	}
	if len(arrSubKey) == 0 {
		return ""
	}
	//按版本从大到小处理查找库文件
	if len(arrSubKey) > 1 {
		sort.SliceStable(arrSubKey, func(i, j int) bool {
			return arrSubKey[i] > arrSubKey[j]
		})
	}
	for _, mPath = range arrSubKey {
		libFile, err := queryRuntimeLib(mPath)
		if err != nil {
			continue
		}
		if _, err = os.Stat(libFile); err == nil {
			return libFile
		}
	}
	return ""
}

func queryRuntimeLib(mPath string) (string, error) {
	var hKey syscall.Handle = 0
	err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, syscall.StringToUTF16Ptr(mPath), 0, syscall.KEY_READ, &hKey)
	if err != nil {
		return "", err
	}
	defer syscall.RegCloseKey(hKey)
	nameLen := uint32(2048)
	nameBuf := make([]uint16, nameLen)
	err = syscall.RegQueryValueEx(hKey, syscall.StringToUTF16Ptr("RuntimeLib"), nil, nil, (*byte)(unsafe.Pointer(&nameBuf[0])), &nameLen)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(nameBuf), nil
}
