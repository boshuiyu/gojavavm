//#include "stdafx.h"
#include "jvm_interface.h"
#include "jvm_core.hpp"
#include "jvm_helper.hpp"
int RESULT_ITEM_SIZE = sizeof(RESULT_ITEM);
//生成返回错误，需要调用者使用完成后释放
char* newBuffer(const char* pszInfo){
	if(pszInfo == NULL){
		return NULL;
	}
	int nLength = strlen(pszInfo);
	char* pszBuffer = (char*)malloc(nLength+1);
	strncpy(pszBuffer,pszInfo,nLength);
	pszBuffer[nLength] = 0x0;
	return pszBuffer;
}
//接口函数实现
void FreeBuffer(void* p){
	if(p != NULL){
		free(p);
	}
}
void* CreateJavaVM(const char* pszJvmLib,const char* pszJarFile,char** ppError){
	string		mErrMsg = "";
	CJVMCore*	pJvm = new CJVMCore();
	if(!pJvm->CreateJavaVM(pszJvmLib,pszJarFile,mErrMsg)){
		*ppError = newBuffer(mErrMsg.c_str());
		pJvm->DestroyJavaVM();
		delete pJvm;
		return NULL;
	}
	return (void*)pJvm;
}
void	DestroyJavaVM(void* p){
	if(p != NULL){
		CJVMCore* pJVM = (CJVMCore*)p;
		pJVM->DestroyJavaVM();
		delete pJVM;
	}
}
jclass	FindClass(void* p,const char* clsName){
	return ((CJVMCore*)p)->FindClass(clsName);
}
jmethodID	GetMethodID(void* p,jclass clazz,const char *name,const char *sig){
	return ((CJVMCore*)p)->GetMethodID(clazz,name,sig);
}
jmethodID	GetStaticMethodID(void* p,jclass clazz,const char *name,const char *sig){
	return ((CJVMCore*)p)->GetStaticMethodID(clazz,name,sig);
}
jobject	AllocGlobalObject(void* p,jclass clazz){
	return ((CJVMCore*)p)->AllocGlobalObject(clazz);
}
void	DeleteGlobalObject(void* p,jobject obj){
	((CJVMCore*)p)->DeleteGlobalObject(obj);
}
//各调用方法
//void,bool,int,long的方法调用,合并成long的调用方式
jlong CallLongMethod(void* p,jobject jobj,jmethodID jmid,char jtype,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
	CJVMHelper helper(p);
	jlong n = helper.CallJavaLongMethod(jobj,jmid,(PARAM_ITEM*)pParams,nPaSize,jtype,nIsArray);
	return n;
}
const char* CallStringMethod(void* p,jobject jobj,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
	const char* pszReturn = NULL;
	CJVMHelper helper(p);
	JNIEnv* env = helper.GetEnv();
	jstring mLocal = (jstring)helper.CallJavaObjectMethod(jobj,jmid,(PARAM_ITEM*)pParams,nPaSize,nIsArray);
	if(mLocal == NULL){
		return NULL;
	}
	const char* psz = env->GetStringUTFChars(mLocal,NULL);
	pszReturn = newBuffer(psz);
	env->ReleaseStringUTFChars(mLocal,psz);
	return pszReturn;
}

jobject	CallObjectMethod(void* p,jobject jobj,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
	jobject		mGlobal = NULL;
	CJVMHelper	helper(p);
	JNIEnv* env = helper.GetEnv();
	jobject mLocal = helper.CallJavaObjectMethod(jobj,jmid,(PARAM_ITEM*)pParams,nPaSize,nIsArray);
	if(mLocal != NULL){
		mGlobal = env->NewGlobalRef(mLocal);
	}
	return mGlobal;
}
const char* CallStaticStringMethod(void* p,jclass clazz,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
	const char* pszReturn = NULL;
	CJVMHelper helper(p);
	JNIEnv* env = helper.GetEnv();
	jstring mLocal = (jstring)helper.CallStaticObjectMethod(clazz,jmid,(PARAM_ITEM*)pParams,nPaSize,nIsArray);
	if(mLocal == NULL){
		return NULL;
	}
	const char* psz = env->GetStringUTFChars(mLocal,NULL);
	pszReturn = newBuffer(psz);
	env->ReleaseStringUTFChars(mLocal,psz);
	return pszReturn;
}

jobject	CallStaticObjectMethod(void* p,jclass clazz,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
	jobject		mGlobal = NULL;
	CJVMHelper	helper(p);
	JNIEnv* env = helper.GetEnv();
	jobject mLocal = helper.CallStaticObjectMethod(clazz,jmid,(PARAM_ITEM*)pParams,nPaSize,nIsArray);
	if(mLocal != NULL){
		mGlobal = env->NewGlobalRef(mLocal);
	}
	return mGlobal;
}
void* ExpandJObjectArray(void* p,jobject jobjarr,int* pArrSize){
	CJVMHelper		helper(p);
	int				nIndex;
	RESULT_ITEM*	pResult;
	RESULT_ITEM*	pItem;
	int				nCount;
	JNIEnv*			env	= helper.GetEnv();
	if(jobjarr == NULL || !helper.IsObjectArray(jobjarr)){
		*pArrSize = -1;
		return NULL;
	}
	nCount = env->GetArrayLength((jobjectArray)jobjarr);
	*pArrSize = nCount;
	if(*pArrSize == 0){
		return NULL;
	}
	pResult = (RESULT_ITEM*)malloc(sizeof(RESULT_ITEM)*nCount);
	for(nIndex=0;nIndex<nCount;nIndex++){
		pItem = &pResult[nIndex];
		pItem->nType = RETURN_TYPE_NULL;
		pItem->nLength=0;
		pItem->pszData=NULL;
		jobject jobj = env->GetObjectArrayElement((jobjectArray)jobjarr,nIndex);
		if(jobj == NULL){
			continue;
		}
		if(helper.IsString(jobj)){
			int	nLength = env->GetStringUTFLength((jstring)jobj);
			pItem->nType = RETURN_TYPE_STRING;
			pItem->nLength = nLength;
			const char* psrc = env->GetStringUTFChars((jstring)jobj,NULL);
			pItem->pszData = malloc(nLength+1);
			memcpy(pItem->pszData,psrc,nLength);
			((char*)pItem->pszData)[nLength] = 0x0;
			env->ReleaseStringUTFChars((jstring)jobj,psrc);
		}else if(helper.IsByteArray(jobj)){
			int	nLength = env->GetArrayLength((jbyteArray)jobj);
			jbyte* pBuffer = env->GetByteArrayElements((jbyteArray)jobj,NULL);
			pItem->nType = RETURN_TYPE_BYTES;
			pItem->nLength = nLength;
			if(nLength>0){
				pItem->pszData = malloc(nLength);
				memcpy(pItem->pszData,pBuffer,nLength);
			}
			env->ReleaseByteArrayElements((jbyteArray)jobj,pBuffer,0);
		}
		env->DeleteLocalRef(jobj);
	}
	return pResult;
}