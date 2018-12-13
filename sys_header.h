#ifndef SYS_HEADER_H
#define SYS_HEADER_H

#ifdef _WIN32
#include "jni_win/jni.h"
#include "jni_win/jni_md.h"
#else
#include "jni_lnx/jni.h"
#include "jni_lnx/jni_md.h"
#endif

#endif