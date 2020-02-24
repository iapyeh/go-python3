
#ifndef IAP_PATCHED_H
#define IAP_PATCHED_H

#include "Python.h"

PyObject *_go_PyIter_Next(PyObject *o);

extern PyObject *_go_PyGen_Type;
int _go_PyGen_Check(PyObject *o);
int _go_PyGen_CheckExact(PyObject *o);
//int _go_PyTraceBack_Print(PyObject *v, PyObject *f);
// Get builtin function, ex. next()
PyObject *_go_PyBuiltin_Get(const char *name);

//class
typedef struct
{
    PyObject_HEAD
    void *userptr; //pointer to User (in go)
} GoUser;

typedef struct
{
    PyObject_HEAD
    void *ptr; //pointer to fasthttp.RequestCtx.Response.Header
} CtxRespHeader;

typedef struct
{
    PyObject_HEAD
    void *ptr; //pointer to fasthttp.RequestCtx.Response
    CtxRespHeader *header;
} CtxResponse;

typedef struct
{
    PyObject_HEAD
    void *ptr; //pointer to fasthttp.RequestCtx.Request.Header
} CtxReqHeader;

typedef struct
{
    PyObject_HEAD
    void *ptr; //pointer to fasthttp.RequestCtx.Response
    CtxReqHeader *header;
} CtxRequest;

typedef struct
{
    PyObject_HEAD
    void *ctxptr; //pointer to RequestCtx (in go)
    GoUser *user; 
    CtxResponse *response;// wrapper to fasthttp.Response
    CtxRequest *request; //  wrapper to fasthttp.Request
    PyObject *metadata; //some handly data for python quick access
} CtxObject;

void SetCtxPtr(PyObject *, void *, void *, void *, void *, void *, void *);

typedef struct
{
    PyObject_HEAD
    void *ctxptr; //pointer to RequestCtx (in go)
    GoUser *user;
    CtxResponse *response;// wrapper to fasthttp.Response
    CtxRequest *request; //  wrapper to fasthttp.Request
    PyObject *metadata; //some handly data for python quick access
} FileUploadCtxObject;
void SetFileUploadCtxPtr(PyObject *, void *, void *, void *, void *, void *, void *);
void FileUploadSave(PyObject *, PyObject *);

typedef struct
{
    PyObject_HEAD
    void *ctxptr; //pointer to RequestCtx (in go)
    GoUser *user;
    PyObject *metadata; //some handly data for python quick access
} WebsocketCtxObject;

void SetWebsocketCtxPtr(PyObject *, void *, void *);
void WebsocketCtxOnMessage(PyObject *, PyObject *);





// ObjshRouter bridges python script and golang.
// It makes @Router.Websocket(), @Router.Get() be valid in python.
typedef struct
{
    PyObject_HEAD
    //void *routerptr; //pointer to Objsh.Router
} ObjshRouter;

// Tree System starts
typedef struct
{
    PyObject_HEAD
    void *ctxptr; //pointer to TreeCallCtx (in go)
    void *wsctxptr;//pointer to TreeCallCtx.WsCtx
    GoUser *user;
    PyObject *metadata; //some handly data for python quick access
    PyObject *onkill;//callback when task been killed
} TreeCallCtxObject;
void SetTreeCallCtxPtr(PyObject *,void *, void *, void *);

// ObjshTree bridges python script and golang.
// It makes @Tree.UnitTest.addBranch be valid in python.
typedef struct
{
    PyObject_HEAD
    //void *routerptr; //pointer to Objsh.Router
} ObjshTree;


PyMODINIT_FUNC PyInit_objshModule(void);
static PyModuleDef objshmodule;



#endif