#include "jvm_interface.h"
class CJVMHelper
{
private:
	CJVMCore*	core;
	JavaVM*		jvm;
	JNIEnv*		env;
private:
	CJVMHelper(void){}
public:
	CJVMHelper(void* p){
		env = NULL;
		core = (CJVMCore*)p;
		jvm = core->pJavaVM;
		jvm->AttachCurrentThread((void**)&env,NULL);
		env->PushLocalFrame(0);
	}
	~CJVMHelper(void){
		env->PopLocalFrame(NULL);
		jvm->DetachCurrentThread();
	}
	JNIEnv* GetEnv(){
		return env;
	}
public:
	jboolean IsObjectArray(jobject jobj){
		return env->IsInstanceOf(jobj,core->clsObjectArray);
	}
	jboolean IsByteArray(jobject jobj){
		return env->IsInstanceOf(jobj,core->clsByteArray);
	}
	jboolean IsString(jobject jobj){
		return env->IsInstanceOf(jobj,core->clsString);
	}
public:
	//方法调用相关函数
	jlong CallJavaLongMethod(jobject jobj,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,char callType,int nIsArray){
		jlong	nResult = -1;
		jvalue* pArgs	= NULL;
		if(nIsArray ==1){
			pArgs = core->createArrayParam(env,pParams,nPaSize);
		}else{
			pArgs = core->createParam(env,pParams,nPaSize);
		}
		if(callType == 'V'){
			env->CallVoidMethodA(jobj,jmid,pArgs);
			nResult = 1;
		}else if(callType == 'I'){
			nResult = jlong(env->CallIntMethodA(jobj,jmid,pArgs));
		}else if(callType == 'J'){
			nResult = env->CallLongMethodA(jobj,jmid,pArgs);
		}else if(callType == 'Z'){
			if(env->CallBooleanMethodA(jobj,jmid,pArgs)){
				nResult = 1;
			}else{
				nResult = 0;
			}
		}
		if(pArgs != NULL){ free(pArgs); }
		return nResult;
	}
	jobject CallJavaObjectMethod(jobject jobj,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
		jlong	nResult = -1;
		jvalue* pArgs	= NULL;
		if(nIsArray==1){
			pArgs = core->createArrayParam(env,pParams,nPaSize);
		}else{
			pArgs = core->createParam(env,pParams,nPaSize);
		}
		jobject mLocal = env->CallObjectMethodA(jobj,jmid,pArgs);
		if(pArgs != NULL){ free(pArgs); }
		return mLocal;
	}
	jobject CallStaticObjectMethod(jclass clazz,jmethodID jmid,PARAM_ITEM* pParams,int nPaSize,int nIsArray){
		jlong	nResult = -1;
		jvalue* pArgs	= NULL;
		if(nIsArray==1){
			pArgs = core->createArrayParam(env,pParams,nPaSize);
		}else{
			pArgs = core->createParam(env,pParams,nPaSize);
		}
		jobject mLocal = env->CallStaticObjectMethodA(clazz,jmid,pArgs);
		if(pArgs != NULL){ free(pArgs); }
		return mLocal;
	}
};