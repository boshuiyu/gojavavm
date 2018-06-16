// Test.cpp : 定义控制台应用程序的入口点。
//

#include "stdafx.h"
#include "../interface.h"
int _tmain(int argc, _TCHAR* argv[])
{
	void* pJvm = NULL;
	const char*	pszErr = CreateJavaVM(&pJvm,
		"C:\\Program Files\\Java\\jdk1.8.0_71\\jre\\bin\\client\\jvm.dll",
		"Z:\\GOWORK\\GOPATH\\src\\javavm\\gojdbc\\testdriver\\gojdbc.jar"
		);
	DestroyJavaVM(pJvm);
	PARAM_ITEM* params = (PARAM_ITEM*)malloc(sizeof(PARAM_ITEM)*3);
	memset(params,0,sizeof(PARAM_ITEM)*3);
	params[0].nType = PARAM_TYPE_STRING;
	params[1].nType = PARAM_TYPE_STRING;
	params[2].nType = PARAM_TYPE_STRING;
	params[0].pData = "jdbc:mysql://127.0.0.1:3306/dbname?pp=2&pass=123456&user=root";
	params[1].pData = "root";
	params[2].pData = "yw@123456";
	//CallJavaMethodEx(&jvmEx,"Open(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Z",params,3);
	return 0;
}
