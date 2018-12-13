#ifndef JAVA_INTERFACE_H
#define JAVA_INTERFACE_H
#include "sys_header.h"
#ifdef _WIN32
#include <Windows.h>
#include <WinBase.h>
#ifndef _WINDOWS_
#define _WINDOWS_
#endif
#else
#include "dlfcn.h"
#include "errno.h"
#endif
//传入参数类型定义
enum PARAM_TYPE{
	PARAM_TYPE_NULL,
	PARAM_TYPE_STRING,
	PARAM_TYPE_INT64,
	PARAM_TYPE_BYTES,
	PARAM_TYPE_JOBJECT
};
//传入参数的结构体
typedef struct tag_ParamItem{
	int		nType;		//类型
	int		nLength;	//数据长度,PARAM_TYPE_PARAMLIST,BYTE[]才有效
	/*nIntVal复用参数
	PARAM_TYPE_INT64时使用这个存储数值
	PARAM_TYPE_NULL,PARAM_TYPE_STRING,PARAM_TYPE_PARAMLIST,PARAM_TYPE_JOBJECT,存储指针
	*/
	jlong	nIntVal;
}PARAM_ITEM;
//输出数据类型定义,NULL,STRING,BYTES就够了，到了go语音层面会自动转换,数字那些全部用string
enum ARRAY_RESULT_TYPE{
	RETURN_TYPE_NULL,
	RETURN_TYPE_STRING,
	RETURN_TYPE_BYTES
};
typedef struct tag_ResultItem{
	int		nType;		//返回类型
	int		nLength;	//数据长度 
	void*	pszData;	//数据地址
}RESULT_ITEM;

//调用函数结果定义
/*
** Make sure we can call this stuff from C++.
*/
#ifdef __cplusplus
extern "C" {
#endif

#ifndef JVM_EXTERN
# define JVM_EXTERN extern
#endif
//定义返回结构体长度
JVM_EXTERN	int			RESULT_ITEM_SIZE;
//释放内部由于newBuffer生成的内存
JVM_EXTERN	void		FreeBuffer(void* p);
//创建，删除，释放等相关函数
JVM_EXTERN void*		CreateJavaVM(const char* pszJvmLib,const char* pszJarFile,char** ppError);
JVM_EXTERN void			DestroyJavaVM(void* pJVM);
//查找类和函数
JVM_EXTERN jclass		FindClass(void* pJVM,const char* clsName);
JVM_EXTERN jmethodID	GetMethodID(void* pJVM,jclass clazz,const char *name,const char *sig);
JVM_EXTERN jmethodID	GetStaticMethodID(void* pJVM,jclass clazz,const char *name,const char *sig);

//生成释放对象
JVM_EXTERN jobject		AllocGlobalObject(void* pJVM,jclass clazz);
JVM_EXTERN void			DeleteGlobalObject(void* pJVM,jobject obj);
//类函数调用
JVM_EXTERN jlong		CallLongMethod(void* p,jobject jobj,jmethodID jmid,char jtype,PARAM_ITEM* pParams,int nPaSize,int nIsArray);
JVM_EXTERN const char*	CallStringMethod(void* p,jobject jobj,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray);
JVM_EXTERN jobject		CallObjectMethod(void* p,jobject jobj,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray);
//静态函数调用
JVM_EXTERN const char*	CallStaticStringMethod(void* p,jclass clazz,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray);
JVM_EXTERN jobject		CallStaticObjectMethod(void* p,jclass clazz,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray);

JVM_EXTERN void*		ExpandJObjectArray(void* p,jobject jobj,int* pArrSize);
#ifdef __cplusplus
}  /* end of the 'extern "C"' block */
#endif
#endif