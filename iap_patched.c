/*
REF:
https://github.com/skyrings/skyring-common/blob/master/tools/gopy/gopy.go
*/
#include "iap_patched.h"
#include "bltinmodule.h"
#include "_cgo_export.h"
#include "structmember.h" //provide T_OBJECT_EX, T_OBJECT
#include <stdio.h> //print()
PyObject *_go_PyIter_Next(PyObject *o) {
    return PyIter_Next(o);
}

//_go_PyGen_Check will become "C._go_PyGen_Check" in go
// PyGen_Check is in CPython API
int _go_PyGen_Check(PyObject *o){
    return PyGen_Check(o);
}

int _go_PyGen_CheckExact(PyObject *o){
    return PyGen_CheckExact(o);
}

PyObject *_go_PyBuiltin_Get(const char *name){
    PyObject *builtins = PyEval_GetBuiltins(); 
    PyObject *callable = PyDict_GetItemString(builtins , name);
    return callable;
}

//global utilities

PyObject* json_dumps_method = NULL;
static int init_json_dumps() {
    /* Import json.loads */
    PyObject* json_module = PyImport_ImportModule("json");
    if (json_module == NULL) {
        return 0;
    }
    json_dumps_method = PyObject_GetAttrString(json_module, "dumps");
    int ret = json_dumps_method != NULL;
    if (ret != 0) Py_XINCREF(json_dumps_method);
    return ret;
}
char * as_string(PyObject *object) {
    if (object == NULL) {
        return NULL;
    }
    if (!PyUnicode_Check(object)) {
        return NULL;
    }
    //char *retval = PyUnicode_AsUTF8(object);
    //return retval;
    return (char *)PyUnicode_AsUTF8(object);
}


/*
 GoUser
*/

static int
objsh_GoUserInit(GoUser *self, PyObject *args, PyObject *kws)
{
    return 0;
}

static PyObject *
gouser_getuserdata(GoUser *self, PyObject *args){
    PyObject *ret;
    if (PyTuple_Size > 0){
        ret = goGetUserData(self->userptr, args);
        Py_XINCREF(ret);
        return ret;
    }
    Py_RETURN_NONE;
}

static PyMethodDef GoUserMethods[] = {
    {"getUserdata", (PyCFunction) gouser_getuserdata, METH_VARARGS, "test function\n"},
	{NULL},
};

static PyObject *
gouser_getmetadata(GoUser *self, PyObject *args){
    PyObject *ret;
    ret = goUserMetadata(self->userptr);
    Py_XINCREF(ret);
    return ret;
}


static PyObject *
gouser_username(GoUser *self, void *unused) {
    char name[] = "Username";
    if (self->userptr == NULL){
        Py_RETURN_NONE;
    }

    PyObject *ret = goUserGetter(self->userptr,name);
    Py_XINCREF(ret);
    return ret;
};

static PyGetSetDef GoUserGetSet[] = {
    { "username", (getter) gouser_username, NULL,  PyDoc_STR("A getset descriptor") },    
    { "metadata", (getter) gouser_getmetadata, NULL,  PyDoc_STR("A getset descriptor") },    
	{ NULL }
};
static PyTypeObject GoUserType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.User",
    .tp_basicsize = sizeof(GoUser),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods = GoUserMethods,
    .tp_init = (initproc)objsh_GoUserInit,
    .tp_getset = GoUserGetSet,
};


/*
 CtxRespHeader
 Wrap fasthttp.ResponseHeader
 Ref: https://github.com/valyala/fasthttp/blob/85217e0d5e01adcd51e6fe5142fa1bda07eaf50b/header.go
*/

/*
 2019-11-21T04:11:34+00:00
 關於 ctx.request, ctx.response 的物件暫時取消，因為主要的功能，如
 getHeader, setHeader已經在ctx中實作，似乎沒具體的用途

static PyObject *
ctxrespheader_setcontenttype(CtxRespHeader *self, PyObject *args){
    char *typename = "ResponseHeader";
    char *funcname = "SetContentType";
    return goCallFunc(typename, self->ptr,funcname,args);
}
static PyObject *
ctxrespheader_set(CtxRespHeader *self, PyObject *args){
    PyObject *ret;
    Py_BEGIN_ALLOW_THREADS
    char *typename = "ResponseHeader";
    char *funcname = "Set";
    ret = goCallFunc(typename, self->ptr,funcname,args);
    Py_END_ALLOW_THREADS
    Py_XINCREF(ret);
    return ret;
}


static PyGetSetDef CtxRespHeaderGetSet[] = {
    //{ "username", (getter) gouser_username, NULL,  PyDoc_STR("A getset descriptor") },    
	{ NULL }
};

static PyMethodDef CtxRespHeaderMethods[] = {
    {"setContentType", (PyCFunction) ctxrespheader_setcontenttype, METH_VARARGS, "test function\n"},
    {"set", (PyCFunction) ctxrespheader_set, METH_VARARGS, "test function\n"},
	{NULL},
};
static PyMemberDef CtxRespHeaderMembers[] = {
    //mp_ass_subscript
    //{"header", T_OBJECT , offsetof(CtxResponse, header), READONLY ,"header"},
    {NULL} 
};
*/
static int
objsh_CtxRespHeaderInit(CtxRespHeader *self, PyObject *args, PyObject *kws)
{
    return 0;
}
static PyTypeObject CtxRespHeaderType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.CtxRespHeader",
    .tp_basicsize = sizeof(CtxRespHeader),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_init = (initproc)objsh_CtxRespHeaderInit,
    //.tp_methods = CtxRespHeaderMethods,
    //.tp_getset = CtxRespHeaderGetSet,
    //.tp_members = CtxRespHeaderMembers,
};

/*
 CtxResponse
 Wrap fasthttp.Response 
 Ref: https://github.com/valyala/fasthttp/blob/2edabf3b76473af8d82b4a746ae8f3f6fe31dca8/http.go
*/

static int
objsh_CtxResponseInit(CtxResponse *self, PyObject *args, PyObject *kws)
{
    self->header = (CtxRespHeader*) PyObject_CallObject((PyObject *) &CtxRespHeaderType, NULL);
    return 0;
}
/*
static PyGetSetDef CtxResponseGetSet[] = {
    //{ "username", (getter) gouser_username, NULL,  PyDoc_STR("A getset descriptor") },    
	{ NULL }
};

static PyMethodDef CtxResponseMethods[] = {
	{NULL},
};
static PyMemberDef CtxResponseMembers[] = {
    {"header", T_OBJECT , offsetof(CtxResponse, header), READONLY ,"header"},
    {NULL} 
};
*/
static PyTypeObject CtxResponseType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.CtxResponse",
    .tp_basicsize = sizeof(CtxResponse),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_init = (initproc)objsh_CtxResponseInit,
    //.tp_getset = CtxResponseGetSet,
    //.tp_methods = CtxResponseMethods,
    //.tp_members = CtxResponseMembers,
};


/*
 CtxReqHeader
 Wrap fasthttp.RequestHeader
 Ref: https://github.com/valyala/fasthttp/blob/85217e0d5e01adcd51e6fe5142fa1bda07eaf50b/header.go

 2019-11-21T04:11:34+00:00
 關於 ctx.request, ctx.response 的物件暫時取消，因為主要的功能，如
 getHeader, setHeader已經在ctx中實作，似乎沒具體的用途

*/
/*
static PyObject *
ctxreqheader_get(CtxReqHeader *self, PyObject *args){
    char *typename = "RequestHeader";
    char *funcname = "Peek";
    return goCallFunc(typename, self->ptr,funcname,args);
}

static PyGetSetDef CtxReqHeaderGetSet[] = {
    //{ "username", (getter) gouser_username, NULL,  PyDoc_STR("A getset descriptor") },    
	{ NULL }
};

static PyMethodDef CtxReqHeaderMethods[] = {
    {"get", (PyCFunction) ctxreqheader_get, METH_VARARGS, "test function\n"},
    {"peek", (PyCFunction) ctxreqheader_get, METH_VARARGS, "test function\n"},
	{NULL},
};

static PyMemberDef CtxReqHeaderMembers[] = {
    //{"header", T_OBJECT , offsetof(CtxResponse, header), READONLY ,"header"},
    {NULL} 
};
*/

static int
objsh_CtxReqHeaderInit(CtxReqHeader *self, PyObject *args, PyObject *kws)
{
    return 0;
}
static PyTypeObject CtxReqHeaderType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.CtxReqHeader",
    .tp_basicsize = sizeof(CtxReqHeader),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_init = (initproc)objsh_CtxReqHeaderInit,
    //.tp_methods = CtxReqHeaderMethods,
    //.tp_getset = CtxReqHeaderGetSet,
    //.tp_members = CtxReqHeaderMembers,
};

/*
 CtxRequest
 Wrap fasthttp.Request
 Ref: https://github.com/valyala/fasthttp/blob/2edabf3b76473af8d82b4a746ae8f3f6fe31dca8/http.go
*/

static int
objsh_CtxRequestInit(CtxRequest*self, PyObject *args, PyObject *kws)
{
    self->header = (CtxReqHeader*) PyObject_CallObject((PyObject *) &CtxReqHeaderType, NULL);
    return 0;
}

/*
static PyGetSetDef CtxRequestGetSet[] = {
    //{ "username", (getter) gouser_username, NULL,  PyDoc_STR("A getset descriptor") },    
	{ NULL }
};
static PyMethodDef CtxRequestMethods[] = {
	{NULL},
};
static PyMemberDef CtxRequestMembers[] = {
    {"header", T_OBJECT , offsetof(CtxRequest, header), READONLY ,"header"},
    {NULL} 
};
*/
static PyTypeObject CtxRequestType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.CtxRequest",
    .tp_basicsize = sizeof(CtxRequest),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_init = (initproc)objsh_CtxRequestInit,
    //.tp_methods = CtxRequestMethods,
    //.tp_getset = CtxRequestGetSet,
    //.tp_members = CtxRequestMembers,
};

/*
 CtxObject (handles Get and Post)
*/
// Returns a dict of query strings in URL request.
// Usage in python: ctx.Kw()
// Deprecated 2019-11-16T11:45:32+00:00, use peek() instead
/*
static PyObject* 
ctx_getKw(CtxObject *self) {
	
    if (self->metadata == NULL){
        Py_RETURN_NONE;
    }

    PyObject *kw;

    Py_BEGIN_ALLOW_THREADS
	PyGILState_STATE state = PyGILState_Ensure();
	
    const char *key = "kw";
    kw = PyDict_GetItemString(self->metadata,key);

    PyGILState_Release(state);

    Py_END_ALLOW_THREADS

    return kw;
}
*/

// Wrapper to resquest header's peek
// peek(key), peek(key,default-value)
// if key if not presented in query, return None
static PyObject* 
ctx_peek(CtxObject *self, PyObject *args) {
    int size = PyTuple_Size(args);
    if (size == 0 ) Py_RETURN_NONE;
    PyObject *arg = PyTuple_GET_ITEM(args,0);
    if (PyUnicode_Check(arg)){
        PyObject *ret = goCtxPeek(self->ctxptr, arg);
        if (ret != Py_None) {
            return ret;
        }
        else if (size == 1){
            // Dont call return Py_None, it is diffrent from Py_RETURN_NONE;
            // (it causes "Fatal Python error: deallocating None")
            // use Py_RETURN_NONE instead.
            Py_RETURN_NONE;
        }
        else{
            ret = PyTuple_GET_ITEM(args,1);
            // Must retain this value
            Py_INCREF(ret);
            return ret;
        }
    }
    Py_RETURN_NONE;
}

static PyObject* 
ctx_setHeader(CtxObject *self, PyObject *args) {
    PyObject *ret;
    Py_BEGIN_ALLOW_THREADS
    char *typename = "ResponseHeader";
    char *funcname = "Set";
    goCallFunc(typename, self->response->header->ptr,funcname,args);
    Py_END_ALLOW_THREADS
    //Py_XINCREF(ret);
    Py_RETURN_NONE;
}

static PyObject* 
ctx_getHeader(CtxObject *self, PyObject *args) {
    PyObject *ret;
    //Py_BEGIN_ALLOW_THREADS
    char *typename = "RequestHeader";
    char *funcname = "Peek";
    ret = goCallFunc(typename, self->request->header->ptr,funcname,args);
    //Py_END_ALLOW_THREADS
    Py_XINCREF(ret);
    return ret;
}

static PyObject* 
remoteAddr(CtxObject *self) {
	PyObject *ret;
    //Py_BEGIN_ALLOW_THREADS
    ret = goRemoteAddr(self->ctxptr);
    //Py_END_ALLOW_THREADS
    return ret;
}

static PyObject* 
ctx_write(CtxObject *self, PyObject *args, PyObject *kws) {
    // arg :bytes
    Py_BEGIN_ALLOW_THREADS
    PyObject* arg = PyTuple_GET_ITEM(args,0);
    if (PyBytes_Check(arg)){
        goCtxWrite(self->ctxptr,arg);
    }
    Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static PyObject* 
ctx_sendfile(CtxObject *self, PyObject *args, PyObject *kws) {
    // arg: string, file path 
    Py_BEGIN_ALLOW_THREADS
    PyObject* arg = PyTuple_GET_ITEM(args,0);
    if (PyUnicode_Check(arg)){
        goCtxSendfile(self->ctxptr,arg);
    }
    Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static PyObject* 
ctx_redirect(CtxObject *self, PyObject *args, PyObject *kws) {
    // arg: string, file path 
    // arg: int, status code, default to 307 , allowed only 301,302,303,307,308
    // REF: https://godoc.org/github.com/valyala/fasthttp#RequestCtx.Redirect
    Py_BEGIN_ALLOW_THREADS
    PyObject* arg0 = PyTuple_GET_ITEM(args,0);
    if (!PyUnicode_Check(arg0)) {
        // nothing
    }else{
        int status = 307;
        PyObject* arg1 = PyTuple_GET_ITEM(args,1);
        if (PyUnicode_Check(arg1)){
            //convert PyUnicode to C string (*char)
            Py_ssize_t size;
            const char *ptr = PyUnicode_AsUTF8AndSize(arg1, &size);
            if (ptr) {
                // convert c string to PyLong
                arg1 = PyLong_FromString(ptr,NULL,10);
                status = (int) PyLong_AsLong(arg1);
            }
        } else if (PyLong_Check(arg1)){
            status = (int) PyLong_AsLong(arg1);
        }

        goCtxRedirect(self->ctxptr, arg0, status);
    }
    Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static int
objsh_CtxObjectInit(CtxObject *self, PyObject *args, PyObject *kws)
{
    self->metadata =  (PyObject*) PyTuple_GET_ITEM(args,0);
    
    self->request =  (CtxRequest*) PyObject_CallObject((PyObject *) &CtxRequestType, NULL);
    
    self->response =  (CtxResponse*) PyObject_CallObject((PyObject *) &CtxResponseType, NULL);
    /*
    // 在此建立user物件會使得pubicMode時，ctx.user 仍有物件，但 ctx.user.username is None
    PyObject *user;
    user =  PyObject_CallObject((PyObject *) &GoUserType, NULL);
    if (user != NULL){
        self->user = (GoUser*) user;
        Py_XINCREF(user);
    }
    */
    return 0;
}

static PyMethodDef CtxObjectMethods[] = {
	//{"kw", (PyCFunction)ctx_getKw, METH_NOARGS, "test function\n"},
	{"peek", (PyCFunction)ctx_peek, METH_VARARGS, "test function\n"},
	{"setHeader", (PyCFunction)ctx_setHeader, METH_VARARGS, "wrap to response.header.set\n"},
	{"getHeader", (PyCFunction)ctx_getHeader, METH_VARARGS, "wrap to request.header.get\n"},
    {"remoteAddr", (PyCFunction)remoteAddr, METH_NOARGS, "test function\n"},
    {"write", (PyCFunction)ctx_write, METH_VARARGS, "write bytes\n"},
    {"sendfile", (PyCFunction)ctx_sendfile, METH_VARARGS, "send file directly\n"},
    {"redirect", (PyCFunction)ctx_redirect, METH_VARARGS, "request client to redirect\n"},
	{NULL,NULL,0,NULL},
};

static PyMemberDef CtxObjectMembers[] = {
    {"metadata", T_OBJECT , offsetof(CtxObject, metadata), READONLY ,"metadata"},
    {"user", T_OBJECT , offsetof(CtxObject, user), READONLY ,"user"},
    {"request", T_OBJECT , offsetof(CtxObject, request), READONLY ,"response"},
    {"response", T_OBJECT , offsetof(CtxObject, response), READONLY ,"response"},
    {NULL} 
};
//REF: https://github.com/CESNET/libnetconf2/blob/fdc7478e5321c9980fd976be7f58f65332a571ab/python/ssh.c
static PyTypeObject CtxObjectType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.CtxObject",
    .tp_basicsize = sizeof(CtxObject),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods = CtxObjectMethods,
    .tp_init = (initproc)objsh_CtxObjectInit,
    .tp_members = CtxObjectMembers,
};


// Called in go to initialize an instance of CtxObject
void SetCtxPtr(
    PyObject *p, 
    void *ptr, 
    void *reqptr, 
    void *reqheaderptr, 
    void *respptr, 
    void *respheaderptr, 
    void *userptr)
{
    CtxObject *ctx;
    ctx = (CtxObject*) p;
    ctx->ctxptr = ptr;
    ctx->request->ptr = reqptr;
    ctx->request->header->ptr = reqheaderptr;
    ctx->response->ptr = respptr;
    ctx->response->header->ptr = respheaderptr;
    if (userptr){
        // 在此才建立user物件會使得pubicMode時，ctx.user is None
        PyObject *user;
        user =  PyObject_CallObject((PyObject *) &GoUserType, NULL);
        if (user != NULL){
            ctx->user = (GoUser*) user;
            Py_XINCREF(user);
        }
        ctx->user->userptr = userptr;
    }
};


/*
FileUploadCtxObject (is a subclass of  CtxObject)
*/
static PyObject *
fileuploadctx_saveto( FileUploadCtxObject *self, PyObject *args)//, PyObject *kws)
{
    //called by python script to send message
    if (PyTuple_Size(args) > 0){
        long ret;
        Py_BEGIN_ALLOW_THREADS
        PyObject *arg = PyTuple_GET_ITEM(args,0);
        ret = (long) goFileUploadCtxSaveTo(self->ctxptr, arg);
        Py_END_ALLOW_THREADS
        return PyLong_FromLong(ret);
    }
    Py_RETURN_NONE;
}

static int
FileUploadCtxObjectInit( FileUploadCtxObject *self, PyObject *args, PyObject *kws)
{
    self->metadata =  (PyObject*) PyTuple_GET_ITEM(args,0);
    self->request =  (CtxRequest*) PyObject_CallObject((PyObject *) &CtxRequestType, NULL);   
    self->response =  (CtxResponse*) PyObject_CallObject((PyObject *) &CtxResponseType, NULL);
    return 0;
}

static PyMethodDef FileUploadCtxObjectMethods[] = {
    {"saveTo", (PyCFunction)fileuploadctx_saveto, METH_VARARGS, "test function\n"},
	{NULL,NULL,0,NULL},
};

static PyObject *
fileuploadctx_Filename(FileUploadCtxObject *self, void *unused) {
    char name[] = "Filename";
    PyObject *ret = goFileUploadCtxGetter(self->ctxptr,name);
    Py_XINCREF(ret);
    return ret;
};
static PyObject *
fileuploadctx_Filesize(FileUploadCtxObject *self, void *unused){
    char name[] = "Filesize";
    PyObject *ret = goFileUploadCtxGetter(self->ctxptr,name);
    Py_XINCREF(ret);
    return ret;
};

static PyGetSetDef FileUploadCtxObjectGetSet[] = {
    { "filename", (getter)fileuploadctx_Filename, NULL,  PyDoc_STR("A getset descriptor") },    
    { "filesize", (getter)fileuploadctx_Filesize, NULL,  PyDoc_STR("A getset descriptor") },    
	{ NULL }
};
static PyMemberDef FileUploadCtxObjectMembers[] = {
    {NULL} 
};

// FileUploadCtxObjectType is a subclass of RequestCtx
static PyTypeObject FileUploadCtxObjectType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.FileUploadCtxObject",
    .tp_basicsize = sizeof( FileUploadCtxObject),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods =  FileUploadCtxObjectMethods,
    .tp_init = (initproc)FileUploadCtxObjectInit,
    .tp_getset = FileUploadCtxObjectGetSet,
    .tp_members = FileUploadCtxObjectMembers,
};

void SetFileUploadCtxPtr(
    PyObject *p, 
    void *ptr, 
    void *reqptr, 
    void *reqheaderptr, 
    void *respptr, 
    void *respheaderptr, 
    void *userptr)
{
    FileUploadCtxObject *ctx;
    ctx = (FileUploadCtxObject*) p;
    ctx->ctxptr = ptr;
    ctx->request->ptr = reqptr;
    ctx->request->header->ptr = reqheaderptr;
    ctx->response->ptr = respptr;
    ctx->response->header->ptr = respheaderptr;
    if (userptr){
        // 在此才建立user物件會使得pubicMode時，ctx.user is None
        PyObject *user;
        user =  PyObject_CallObject((PyObject *) &GoUserType, NULL);
        if (user != NULL){
            ctx->user = (GoUser*) user;
            Py_XINCREF(user);
        }
        ctx->user->userptr = userptr;
    }
};

/*
Websocket
*/

static PyObject *
websocketctx_on(WebsocketCtxObject *self, PyObject *args, PyObject *kws)
{
    //Called by python script to register event listeners
    PyObject *eventname;
    PyObject *callback;
    eventname =  (PyObject*) PyTuple_GET_ITEM(args,0);
  
    char *cname = (char*) PyUnicode_DATA(eventname);
    callback = (PyObject*) PyTuple_GET_ITEM(args,1);
    
    const char *message = "message";
    const char *close = "close";
    if (strcmp(cname, message) == 0) {
        //self->onmessage = callback;
        //Py_XINCREF(self->onmessage);
        char *evtname = "Message";
        char *token4remove = "py-message-handler";
        goWsCtxAddEventListener(self->ctxptr, evtname, token4remove,callback);
        Py_XINCREF(callback);

    }else if(strcmp(cname, close) == 0){
        char *evtname = "Close";
        char *token4remove = "py-close-handler";
        goWsCtxAddEventListener(self->ctxptr, evtname, token4remove,callback);
        Py_XINCREF(callback);
    }else{
        goPrint("=============3");
    }
    
   Py_RETURN_NONE;
}
static PyObject*
websocketctx_send(WebsocketCtxObject *self, PyObject *args, PyObject *kws)
{
    //called by python script to send message
    if (PyTuple_Size(args) > 0){
        Py_BEGIN_ALLOW_THREADS
        PyObject *arg = PyTuple_GET_ITEM(args,0);
        goWebsocketCtxSend(self->ctxptr, arg);
        Py_END_ALLOW_THREADS
    }
    Py_RETURN_NONE;
}

static int
objsh_WebsocketCtxObjectInit(WebsocketCtxObject *self, PyObject *args, PyObject *kws)
{
    self->metadata =  (PyObject*) PyTuple_GET_ITEM(args,0);
    return 0;
}

static PyMethodDef WebsocketCtxObjectMethods[] = {
    {"on", (PyCFunction)websocketctx_on, METH_VARARGS, "test function\n"},
    {"send", (PyCFunction)websocketctx_send, METH_VARARGS, "test function\n"},
	{NULL,NULL,0,NULL},
};
static PyMemberDef WebsocketCtxObjectMembers[] = {
    {"user", T_OBJECT , offsetof(WebsocketCtxObject, user), READONLY ,"user"},
    {NULL} 
};
static PyTypeObject WebsocketCtxObjectType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.WebsocketCtxObject",
    .tp_basicsize = sizeof(WebsocketCtxObject),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods = WebsocketCtxObjectMethods,
    .tp_init = (initproc)objsh_WebsocketCtxObjectInit,
    .tp_members = WebsocketCtxObjectMembers,
};

// Called in go to initialize an instance of CtxObject
void SetWebsocketCtxPtr(PyObject *p, void *ptr, void *userptr){
    WebsocketCtxObject *wsctx;
    wsctx = (WebsocketCtxObject*) p;
    wsctx->ctxptr = ptr;

    if (userptr){
        // 在此才建立user物件會使得pubicMode時，ctx.user is None
        PyObject *user;
        user =  PyObject_CallObject((PyObject *) &GoUserType, NULL);
        if (user != NULL){
            wsctx->user = (GoUser*) user;
            Py_XINCREF(user);
        }
        wsctx->user->userptr = userptr;
    }

};

//called in golang
/*
void WebsocketCtxOnMessage(PyObject *p, PyObject *mesg){
    WebsocketCtxObject *wsctx;
    wsctx = (WebsocketCtxObject*) p;
    if (wsctx->onmessage != NULL) {
        PyObject_CallFunctionObjArgs(wsctx->onmessage,mesg,NULL);
    }
}
*/

/*
 Wrapper to objsh.Router
 */

 
static PyObject*
objshrouter_Get(ObjshRouter *self, PyObject *args)
{
    //path,handler,acl mode,
    if (PyTuple_Size(args) > 2){
        
        // 在Go當中會require GIL,此處只是一開始時註冊handler而已，不在此鎖GIL應不至於會產生問題
        // see iap_patch.go
        //  func Get(urlpath string, handler *PyObject, acl int, post bool) {
        //Py_BEGIN_ALLOW_THREADS
        //PyGILState_STATE state = PyGILState_Ensure();
        
        PyObject *path = PyTuple_GET_ITEM(args,0);
        PyObject *handler = PyTuple_GET_ITEM(args,1);
        long acl = PyLong_AsLong(PyTuple_GET_ITEM(args,2));
        Py_XINCREF(handler);
        goObjshRouterGet(path,handler,(int)acl);
        // Dont call Py_XDECREF(handler);
        //Py_XDECREF(handler);
        
        //PyGILState_Release(state);        
        //Py_END_ALLOW_THREADS
    }
    Py_RETURN_NONE;
}

static PyObject*
objshrouter_Post(ObjshRouter *self, PyObject *args)
{
    //path,handler,acl mode
    if (PyTuple_Size(args) > 2){
        PyObject *path = PyTuple_GET_ITEM(args,0);
        PyObject *handler = PyTuple_GET_ITEM(args,1);
        long acl = PyLong_AsLong(PyTuple_GET_ITEM(args,2));
        Py_XINCREF(handler);
        goObjshRouterPost(path,handler,(int)acl);
    }
    Py_RETURN_NONE;
}

static PyObject*
objshrouter_Websocket(ObjshRouter *self, PyObject *args)
{
    //path,handler,acl mode
    if (PyTuple_Size(args) > 2){
        PyObject *path = PyTuple_GET_ITEM(args,0);
        PyObject *handler = PyTuple_GET_ITEM(args,1);
        long acl = PyLong_AsLong(PyTuple_GET_ITEM(args,2));
        Py_XINCREF(handler);
        goObjshRouterWebsocket(path,handler,(int)acl);
    }
    Py_RETURN_NONE;
}

static PyObject*
objshrouter_FileUpload(ObjshRouter *self, PyObject *args)
{
    //path,handler,acl mode
    if (PyTuple_Size(args) > 2){
        PyObject *path = PyTuple_GET_ITEM(args,0);
        PyObject *handler = PyTuple_GET_ITEM(args,1);
        long acl = PyLong_AsLong(PyTuple_GET_ITEM(args,2));
        Py_XINCREF(handler);
        goObjshRouterFileUpload(path,handler,(int)acl);
    }
    Py_RETURN_NONE;
}

static int
objsh_ObjshRouterInit(ObjshRouter *self, PyObject *args, PyObject *kws)
{
    return 0;
}

static PyMethodDef ObjshRouterMethods[] = {
    {"Get", (PyCFunction)objshrouter_Get, METH_VARARGS, "test function\n"},
    {"Post", (PyCFunction)objshrouter_Post, METH_VARARGS, "test function\n"},
    {"Websocket", (PyCFunction)objshrouter_Websocket, METH_VARARGS, "test function\n"},
    {"FileUpload", (PyCFunction)objshrouter_FileUpload, METH_VARARGS, "test function\n"},
	{NULL,NULL,0,NULL},
};

static  PyTypeObject ObjshRouterType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.ObjshRouter",
    .tp_basicsize = sizeof(ObjshRouter),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods = ObjshRouterMethods,
    .tp_init = (initproc)objsh_ObjshRouterInit,
};

/*
 * ObjshTree (been called in iap_patched.go's fastjob module)
 */
static PyObject* objshtree_AddBranch(ObjshTree *self, PyObject *args)
{
    /* method to raise error
    if (treeName == NULL){
        PyErr_SetString(PyExc_ValueError,"tree name not existed");
        return NULL;
    }
    */

    Py_BEGIN_ALLOW_THREADS
    // instance of BaseBranch()
    PyObject *branchObj = PyTuple_GET_ITEM(args,0);
    PyObject *treeName = PyTuple_GET_ITEM(args,1);
    Py_XINCREF(branchObj);
    //Py_XINCREF(treeName);
    goObjshTreeAddBranch(branchObj,treeName);

    Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static PyObject* objshtree_SureReady(ObjshTree *self, PyObject *args)
{
    //args = (treename, branch instance,acl mode)
    if (PyTuple_Size(args) > 2){
        PyGILState_STATE state = PyGILState_Ensure();
        Py_BEGIN_ALLOW_THREADS //miso
        PyObject *treeName = PyTuple_GET_ITEM(args,0);
        // instance of BaseBranch()
        PyObject *branchName = PyTuple_GET_ITEM(args,1);
        goObjshTreeSureReady(treeName,branchName);
        Py_END_ALLOW_THREADS
        PyGILState_Release(state);
    }
    Py_RETURN_NONE;
}
static PyMethodDef ObjshTreeMethods[] = {
    {"AddBranch", (PyCFunction)objshtree_AddBranch, METH_VARARGS, "test function\n"},
    {"SureReady", (PyCFunction)objshtree_SureReady, METH_VARARGS, "test function\n"},
	{NULL,NULL,0,NULL},
};
static int objsh_ObjshTreeInit(ObjshTree *self, PyObject *args, PyObject *kws)
{
    return 0;
}
static  PyTypeObject ObjshTreeType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.ObjshTree",
    .tp_basicsize = sizeof(ObjshTree),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods = ObjshTreeMethods,
    .tp_init = (initproc)objsh_ObjshTreeInit,
};

//TreeCallCtxObject
static PyObject *
treecallctx_resolve(TreeCallCtxObject *self, PyObject *args)
{

    //called by python script //miso

    // 2019-09-21T12:23:17+00:00 似乎只要 PyGILState_Ensure，
    // 就不需要 Py_BEGIN_ALLOW_THREADS與 Py_END_ALLOW_THREADS
    //Py_BEGIN_ALLOW_THREADS
 
    PyGILState_STATE state = PyGILState_Ensure();
    char *ret;
    int size = PyTuple_Size(args);
    PyObject *jsonstr;
    if (size == 0){
        ret = "";
    }
    else if (size == 1){
        jsonstr = PyObject_CallObject(json_dumps_method,args);
        ret = as_string(jsonstr);
    }
    else{
        PyObject* onearg_tuple = PyTuple_New(1);
        PyTuple_SET_ITEM(onearg_tuple,0,args);
        jsonstr = PyObject_CallObject(json_dumps_method,onearg_tuple);
        Py_XDECREF(onearg_tuple);
        ret = as_string(jsonstr);
    }
    goTreeCallCtxResolve(self->ctxptr,ret,strlen(ret));
    PyGILState_Release(state);

    //Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static PyObject *
treecallctx_notify(TreeCallCtxObject *self, PyObject *args)
{
    PyGILState_STATE state = PyGILState_Ensure();
    char *ret;
    int size = PyTuple_Size(args);
    PyObject *jsonstr;
    if (size == 0){
        ret = "";
    }
    else if (size == 1){
        jsonstr = PyObject_CallObject(json_dumps_method,args);
        ret = as_string(jsonstr);
        goTreeCallCtxNotify(self->ctxptr,ret);
    }
    else{
        PyObject* onearg_tuple = PyTuple_New(1);
        PyTuple_SET_ITEM(onearg_tuple,0,args);
        jsonstr = PyObject_CallObject(json_dumps_method,onearg_tuple);
        Py_XDECREF(onearg_tuple);
        ret = as_string(jsonstr);
    }
    goTreeCallCtxNotify(self->ctxptr,ret);
    // Py_XDECREF 不能隨便呼叫，似乎對於promary type，叫了會當掉或亂取值
    //Py_XDECREF(ret);
    PyGILState_Release(state);

    //Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static PyObject *
treecallctx_reject(TreeCallCtxObject *self, PyObject *args)
{
    //called by python script
    PyGILState_STATE state = PyGILState_Ensure();
    char *errstr;
    long retcode = 1;
    int size = PyTuple_Size(args);
    PyObject *jsonstr;
    if (size == 0){
        errstr = "";
    }
    else if (size == 1){
        retcode = PyLong_AsLong(PyTuple_GET_ITEM(args,0));
        errstr = "";
    }
    else{
        retcode = PyLong_AsLong(PyTuple_GET_ITEM(args,0));
        PyObject* onearg_tuple = PyTuple_New(1);
        PyObject *err = PyTuple_GET_ITEM(args,1);
        PyTuple_SET_ITEM(onearg_tuple,0,err);
        jsonstr = PyObject_CallObject(json_dumps_method,onearg_tuple);
        errstr = as_string(jsonstr);
    }
    goTreeCallCtxReject(self->ctxptr,retcode,errstr);
    /*
    暫時取消，local var 可能不需要release
    //不能隨便呼叫，只是int,字串而已的話，叫了會當掉。
    //可是如果onearg_tuple放的是全部的args,而不是部分的item (PyTuple_GET_ITEM(args,1))
    //像是在resolve那樣，卻又沒問題(WHY?)
    if (PyDict_Check(err) || PyList_Check(err) || PyTuple_Check(err)){
        Py_XDECREF(onearg_tuple);
        Py_XDECREF(errstr);
    }
    */
    PyGILState_Release(state);

    //Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static PyObject *
treecallctx_setbackground(TreeCallCtxObject *self)
{
    //called by python script
    Py_BEGIN_ALLOW_THREADS
    goTreeCallCtxSetBackground(self->ctxptr);
    Py_END_ALLOW_THREADS
    Py_RETURN_NONE;

}
/* Example of setter
int treecallctx_setonkill(TreeCallCtxObject *self, PyObject *value, void *closure)
{
    self->_onkill = value;
    return 0;
}
*/
static PyObject *
treecallctx_kill(TreeCallCtxObject *self)
{
    //called by golang to notify python script that it is killed
    Py_BEGIN_ALLOW_THREADS
    const char *onkillstr = "onkill";
    PyObject *onkill = PyObject_GetAttrString((PyObject*)self,onkillstr);
    if (onkill != NULL && PyFunction_Check(onkill)){
        // requires GIL
        PyGILState_STATE state = PyGILState_Ensure();
        PyObject_CallObject(onkill,NULL);
        PyGILState_Release(state);
    }
    Py_END_ALLOW_THREADS
    Py_RETURN_NONE;
}
static int
objsh_TreeCallCtxObjectInit(TreeCallCtxObject *self, PyObject *args, PyObject *kws)
{
    self->metadata =  (PyObject*) PyTuple_GET_ITEM(args,0);
    return 0;
}
static PyMethodDef TreeCallCtxObjectMethods[] = {
    {"kill", (PyCFunction)treecallctx_kill,  METH_NOARGS, "test function\n"},
    {"resolve", (PyCFunction)treecallctx_resolve, METH_VARARGS, "test function\n"},
    {"reject", (PyCFunction)treecallctx_reject, METH_VARARGS, "test function\n"},
    {"notify", (PyCFunction)treecallctx_notify, METH_VARARGS, "test function\n"},
    {"setBackground", (PyCFunction)treecallctx_setbackground, METH_NOARGS, "test function\n"},
	{NULL,NULL,0,NULL},
};
static PyObject * treecallctx_getargs(TreeCallCtxObject *self, void *unused) {
    //metadata
    const char *key = "Args";
    PyObject *args = PyDict_GetItemString(self->metadata,key);
    Py_XINCREF(args);
    return args;
};
static PyObject * treecallctx_getkw(TreeCallCtxObject *self, void *unused) {
    //metadata
    const char *key = "Kw";
    PyObject *kw = PyDict_GetItemString(self->metadata,key);
    Py_XINCREF(kw);
    return kw;
};
static PyGetSetDef TreeCallCtxObjectGetSet[] = {
    //{ "onkill", NULL,(setter)treecallctx_setonkill, PyDoc_STR("A getset descriptor") },    
    { "args", (getter)treecallctx_getargs,NULL, PyDoc_STR("A getset descriptor") },    
    { "kw", (getter)treecallctx_getkw,NULL, PyDoc_STR("A getset descriptor") },    
	{ NULL }
};
static PyMemberDef TreeCallCtxObjectMembers[] = {
    {"user", T_OBJECT , offsetof(TreeCallCtxObject, user), READONLY ,"user"},
    {"onkill", T_OBJECT , offsetof(TreeCallCtxObject, onkill), 0 ,"onkill"}, //writable
    {NULL} 
};
static PyTypeObject TreeCallCtxObjectType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "cfastjob.TreeCallCtxObject",
    .tp_basicsize = sizeof( TreeCallCtxObject),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = PyType_GenericNew,
    .tp_methods =  TreeCallCtxObjectMethods,
    .tp_init = (initproc)objsh_TreeCallCtxObjectInit,
    .tp_getset = TreeCallCtxObjectGetSet,
    .tp_members = TreeCallCtxObjectMembers,
};
// Called in go to initialize an instance of CtxObject
void SetTreeCallCtxPtr(PyObject *p, void *ptr, void *wsctxptr, void *userptr){
    TreeCallCtxObject *ctx;
    ctx = (TreeCallCtxObject*) p;
    ctx->ctxptr = ptr;
    ctx->wsctxptr = wsctxptr;
    
    if (userptr){
        // 在此才建立user物件會使得pubicMode時，ctx.user is None
        PyObject *user;
        user =  PyObject_CallObject((PyObject *) &GoUserType, NULL);
        if (user != NULL){
            ctx->user = (GoUser*) user;
            Py_XINCREF(user);
        }
        ctx->user->userptr = userptr;
    }    
};
//Module related starts
static PyModuleDef objshmodule = {
    PyModuleDef_HEAD_INIT,
    .m_name = "cfastjob",
    .m_doc = "module that bridge python and go",
    .m_size = -1,
};


//This will get the noddy module
PyMODINIT_FUNC
PyInit_objshModule(void)
{
    PyObject* m;

    if ( init_json_dumps() == 0)
        return NULL;

    if (PyType_Ready(&GoUserType) < 0)
        return NULL;

    if (PyType_Ready(&CtxReqHeaderType) < 0)
        return NULL;

    if (PyType_Ready(&CtxRequestType) < 0)
        return NULL;

    if (PyType_Ready(&CtxRespHeaderType) < 0)
        return NULL;

    if (PyType_Ready(&CtxResponseType) < 0)
        return NULL;

    if (PyType_Ready(&CtxObjectType) < 0)
        return NULL;

    if (PyType_Ready(&WebsocketCtxObjectType) < 0)
        return NULL;

    FileUploadCtxObjectType.tp_base = &CtxObjectType;
    if (PyType_Ready(&FileUploadCtxObjectType) < 0)
        return NULL;

    if (PyType_Ready(&ObjshRouterType) < 0)
        return NULL;

    if (PyType_Ready(&ObjshTreeType) < 0)
        return NULL;

    if (PyType_Ready(&TreeCallCtxObjectType) < 0)
        return NULL;

    m = PyModule_Create(&objshmodule);
    if (m == NULL)
        return NULL;

    Py_INCREF(&CtxObjectType);
    PyModule_AddObject(m, "CtxObject", (PyObject *) &CtxObjectType);

    Py_INCREF(&WebsocketCtxObjectType);
    PyModule_AddObject(m, "WebsocketCtxObject", (PyObject *) &WebsocketCtxObjectType);

    Py_INCREF(&FileUploadCtxObjectType);
    PyModule_AddObject(m, "FileUploadCtxObject", (PyObject *) &FileUploadCtxObjectType);

    Py_INCREF(&ObjshRouterType);
    PyModule_AddObject(m, "ObjshRouter", (PyObject *) &ObjshRouterType);

    Py_INCREF(&ObjshTreeType);
    PyModule_AddObject(m, "ObjshTree", (PyObject *) &ObjshTreeType);

    Py_INCREF(&TreeCallCtxObjectType);
    PyModule_AddObject(m, "TreeCallCtxObject", (PyObject *) &TreeCallCtxObjectType);

    return m;
}


// retunn an instance;
// 這個不work,因為產生的instance叫不到method,除非先 dir(instance)
// 可能是使用 PyObject_New 的緣故
/*
PyObject*
new_noddy(PyObject *ctx)
{
    
	CtxObject *self;
	self = PyObject_New(CtxObject, &CtxObjectType);
	if (self == NULL)
		return NULL;
    
    //init values here
	self->ctx = ctx;

    return (PyObject*)self;
}
*/