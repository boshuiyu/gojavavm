#include "jvm_interface.h"
#include <vector>
#include <string>
using namespace std;
//jvm相关函数定义
typedef jint (JNICALL *PFN_JNI_CreateJavaVM)(JavaVM **pvm, void **penv, void *args);
//jvm核心类
class CJVMCore
{
public:
	JavaVM*		pJavaVM;
public:
	jclass		clsObject;
	jclass		clsInteger;
	jclass		clsLong;
	jclass		clsString;
	jclass		clsObjectArray;
	jclass		clsByteArray;
	jmethodID	midInitInteger;
	jmethodID	midInitLong;
public:
	CJVMCore(void){
		pJavaVM = NULL;
		clsObject = NULL;
		clsInteger = NULL;
		clsLong = NULL;
		midInitInteger = NULL;
		midInitLong = NULL;
	}
	~CJVMCore(void){
		DestroyJavaVM();
	}

private:
	//读取常用的类和构造函数，后续方便使用
	bool loadCommonObject(JNIEnv* pEnv){
		pEnv->PushLocalFrame(0);
		clsObject = (jclass)pEnv->NewGlobalRef(pEnv->FindClass("java/lang/Object"));
		clsInteger = (jclass)pEnv->NewGlobalRef(pEnv->FindClass("java/lang/Integer"));
		clsLong = (jclass)pEnv->NewGlobalRef(pEnv->FindClass("java/lang/Long"));
		clsString = (jclass)pEnv->NewGlobalRef(pEnv->FindClass("java/lang/String"));
		clsObjectArray = (jclass)pEnv->NewGlobalRef(
			pEnv->GetObjectClass(pEnv->NewObjectArray(0,clsObject,NULL)));
		clsByteArray = (jclass)pEnv->NewGlobalRef(
			pEnv->GetObjectClass(pEnv->NewByteArray(0)));
		if(clsObject == NULL || clsInteger == NULL || clsLong == NULL || clsString == NULL
			||clsObjectArray == NULL || clsByteArray == NULL){
			pEnv->PopLocalFrame(NULL);
			return false;
		}
		midInitInteger = (jmethodID)pEnv->NewGlobalRef((jobject)pEnv->GetMethodID(clsInteger,"<init>","(I)V"));
		midInitLong = (jmethodID)pEnv->NewGlobalRef((jobject)pEnv->GetMethodID(clsLong,"<init>","(J)V"));
		if(midInitInteger == NULL || midInitLong == NULL){
			pEnv->PopLocalFrame(NULL);
			return false;
		}
		pEnv->PopLocalFrame(NULL);
		return true;
	}
	jclass GetObjectArrayClass(JNIEnv* pEnv){
		jclass mLocal = pEnv->GetObjectClass(pEnv->NewObjectArray(0,clsObject,NULL));
		jclass mGlobal = (jclass)pEnv->NewGlobalRef(mLocal);
		return mGlobal;
	}
	//获得，释放环境指针
	JNIEnv* AttachThread(){
		JNIEnv* pEnv = NULL;
		pJavaVM->AttachCurrentThread((void**)&pEnv,NULL);
		return pEnv;
	}
	void	DetachThread(){
		if(pJavaVM != NULL){
			pJavaVM->DetachCurrentThread();
		}
	}
	//JAVA函数创建参数
	int		getParamItemCount(PARAM_ITEM* pParams){
		int			nCount = 0;
		PARAM_ITEM* pItem = pParams;
		while(pItem != NULL){
			nCount++;
			pItem++;
		}
		return nCount;
	}
	
public:
	jvalue* createArrayParam(JNIEnv* pEnv,PARAM_ITEM* pParams,int nCount){
		PARAM_ITEM*		pItem		= pParams;
		int				nIndex		= 0;
		jvalue*			pArgs		= NULL;
		jobjectArray	arrObject	= pEnv->NewObjectArray(nCount,clsObject,NULL);
		for(nIndex =0; nIndex<nCount; nIndex++)
		{
			pItem = &pParams[nIndex];
			switch(pParams[nIndex].nType)
			{
			case PARAM_TYPE_STRING:
				if(pItem->nIntVal != 0){
					pEnv->SetObjectArrayElement(arrObject,nIndex,
						pEnv->NewStringUTF((const char*)pItem->nIntVal));
				}
				break;
			case PARAM_TYPE_INT64:
				pEnv->SetObjectArrayElement(arrObject,nIndex,
					pEnv->NewObject(clsLong,midInitLong,jlong(pItem->nIntVal)));
				break;
			case PARAM_TYPE_BYTES:
				if(pItem->nIntVal != 0 && pItem->nLength > 0){
					jsize nJLen = jsize(pItem->nLength);
					jbyteArray jbtArr =  pEnv->NewByteArray(nJLen);
					pEnv->SetByteArrayRegion(jbtArr,0,nJLen,(jbyte*)pItem->nIntVal);
					pEnv->SetObjectArrayElement(arrObject,nIndex,jbtArr);
				}
				break;
			default:
				//就不用设置了，默认就是null;
				break;
			}
		}
		pArgs = (jvalue*)malloc(sizeof(jvalue));
		memset(pArgs,0,sizeof(jvalue));
		pArgs->l = arrObject;
		return pArgs;
	}
	jvalue* createParam(JNIEnv* pEnv,PARAM_ITEM* pParams,int nCount){
		if(nCount <= 0 || pParams == NULL){
			return NULL;
		}
		PARAM_ITEM* pItem = pParams;
		jvalue*		pArgs = (jvalue*)malloc(sizeof(jvalue)*nCount);
		memset(pArgs,0,sizeof(jvalue)*nCount);
		int			nIndex = 0;
		for(nIndex=0;nIndex<nCount;nIndex++){
			pItem = &pParams[nIndex];
			switch(pItem->nType)
			{
			case PARAM_TYPE_STRING:
				if(pItem->nIntVal == 0){
					pArgs[nIndex].l = NULL;
				}else{
					pArgs[nIndex].l = pEnv->NewStringUTF((const char*)pItem->nIntVal);
				}
				break;
			case PARAM_TYPE_INT64:
				pArgs[nIndex].j = jlong(pItem->nIntVal);
				break;
			case PARAM_TYPE_BYTES:
				if(pItem->nIntVal != 0 && pItem->nLength > 0){
					jsize nJLen = jsize(pItem->nLength);
					jbyteArray jbtArr =  pEnv->NewByteArray(nJLen);
					pEnv->SetByteArrayRegion(jbtArr,0,nJLen,(jbyte*)pItem->nIntVal);
					pArgs[nIndex].l = jbtArr;
				}else{
					pArgs[nIndex].l = NULL;
				}
				break;
			default:
				pArgs[nIndex].l = NULL;
				break;
			}
		}
		return pArgs;
	}
public:
	//获得错误信息
	static string FormatString(const char* fmt,...){
		char szBuffer[1024] = {0};
		va_list argList;
		va_start(argList,fmt);
		if(vsnprintf(&szBuffer[0],1000,fmt,argList)<=0){
			va_end(argList);
			return "unknown error:call vsnprintf";
		}
		va_end(argList);
		return string(szBuffer);
	}
	//创建jvm虚拟机
	bool CreateJavaVM(const char* pszJvmLib,const char* pszJarFile,string& mErrMsg){
		JNIEnv*	pJavaEnv= NULL;
		jint	jResult = 0;
		const int nOptionCount = 3;
		JavaVMOption vmOption[nOptionCount];
		char szJarParamBuffer[2048] = {0};
		PFN_JNI_CreateJavaVM pfnJNI_CreateJavaVM = NULL;
		//准备jvm的初始化参数
		sprintf(szJarParamBuffer,"-Djava.class.path=%s",pszJarFile);
		//设置JVM最大允许分配的堆内存，按需分配  
		vmOption[0].optionString = (char*)"-Xmx256M";  
		vmOption[1].optionString = szJarParamBuffer;  
		vmOption[2].optionString = (char*)"-Djava.compiler=NONE";
		JavaVMInitArgs vmInitArgs;  
		vmInitArgs.version = JNI_VERSION_1_6;  
		vmInitArgs.options = vmOption;  
		vmInitArgs.nOptions = nOptionCount;
#ifdef _WIN32
		HMODULE hModule = LoadLibrary(pszJvmLib);
		if(hModule == NULL){
			mErrMsg = FormatString("CreateJavaVM LoadLibrary Error:%d",GetLastError());
			return false;
		}
		pfnJNI_CreateJavaVM = (PFN_JNI_CreateJavaVM)GetProcAddress(hModule,"JNI_CreateJavaVM");
		if(pfnJNI_CreateJavaVM == NULL){
			mErrMsg = "CreateJavaVM Lost Function:JNI_CreateJavaVM";
			return false;
		}
#else
		return FormatString="CreateJavaVM unsupport BBB(LINUX)";
#endif
		jResult = pfnJNI_CreateJavaVM(&pJavaVM,(void**)&pJavaEnv,&vmInitArgs);
		if(jResult != 0 || pJavaVM == NULL){
			mErrMsg = FormatString("CreateJavaVM Return:%d",int(jResult));
			return false;
		}
		if(!loadCommonObject(pJavaEnv)){
			DestroyJavaVM();
			mErrMsg = "CreateJavaVM Load Common Object Error";
			return false;
		}
		return true;
	}
	//释放jvm虚拟机,一般不用调用
	void		DestroyJavaVM(){
		if(pJavaVM == NULL){
			return;
		}
		JNIEnv* pEnv = AttachThread();
		if(clsObject != NULL){ 
			pEnv->DeleteGlobalRef(clsObject);
			clsObject = NULL;
		}
		if(clsInteger != NULL){ 
			pEnv->DeleteGlobalRef(clsInteger);
			clsInteger = NULL;
		}
		if(clsLong != NULL){ 
			pEnv->DeleteGlobalRef(clsLong);
			clsLong = NULL;
		}
		if(midInitInteger != NULL){ 
			pEnv->DeleteGlobalRef((jobject)midInitInteger);
			midInitInteger = NULL;
		}
		if(midInitLong != NULL){ 
			pEnv->DeleteGlobalRef((jobject)midInitLong);
			midInitLong = NULL;
		}
		DetachThread();
		pJavaVM->DestroyJavaVM();
		pJavaVM = NULL;
	}
	//查找java类
	jclass		FindClass(const char* clsName){
		JNIEnv* pEnv = AttachThread();
		jclass  mLocal = pEnv->FindClass(clsName);
		if(mLocal == NULL){
			return NULL;
		}
		jclass  mGlobal = (jclass)pEnv->NewGlobalRef(mLocal);
		pEnv->DeleteLocalRef(mLocal);
		DetachThread();
		return  mGlobal;
	}
	//查找java类的方法
	jmethodID	GetMethodID(jclass clazz,const char *name,const char *sig){
		JNIEnv* pEnv = AttachThread();
		jmethodID  mLocal = pEnv->GetMethodID(clazz,name,sig);
		if(mLocal == NULL){
			return NULL;
		}
		jmethodID  mGlobal = (jmethodID)pEnv->NewGlobalRef((jobject)mLocal);
		pEnv->DeleteLocalRef((jobject)mLocal);
		DetachThread();
		return mGlobal;
	}
	jmethodID	GetStaticMethodID(jclass clazz,const char *name,const char *sig){
		JNIEnv* pEnv = AttachThread();
		jmethodID  mLocal = pEnv->GetStaticMethodID(clazz,name,sig);
		if(mLocal == NULL){
			return NULL;
		}
		jmethodID  mGlobal = (jmethodID)pEnv->NewGlobalRef((jobject)mLocal);
		pEnv->DeleteLocalRef((jobject)mLocal);
		DetachThread();
		return mGlobal;
	}
	//生成一个类的全局引用实例,默认构造函数
	jobject	AllocGlobalObject(jclass clazz){
		JNIEnv* pEnv = AttachThread();
		jobject  mLocal = pEnv->AllocObject(clazz);
		if(mLocal == NULL){
			return NULL;
		}
		jobject  mGlobal = pEnv->NewGlobalRef(mLocal);
		pEnv->DeleteLocalRef(mLocal);
		DetachThread();
		return mGlobal;
	}
	//删除一个类的全局引用实例
	void DeleteGlobalObject(jobject obj){
		JNIEnv* pEnv = AttachThread();
		pEnv->DeleteGlobalRef(obj);
		DetachThread();
	}
};
