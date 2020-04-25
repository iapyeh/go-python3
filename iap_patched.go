package python3

/*
There should be a python3.c in your pkgconfig path, ex. 
/usr/lib/x86_64-linux-gnu/pkgconfig/ in Ubuntu, if not make a symblic link for it.
For example, ln -s python-3.7.pc python3.pc to create it.
In MacOS, the path might be /usr/local/lib/pkgconfig
*/

/*
#cgo !windows pkg-config: python3
#include "Python.h"
#include "iap_patched.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/valyala/fasthttp"

	"github.com/iapyeh/fastjob/model"
	"github.com/mattn/go-pointer"
)

type RequestCtx = model.RequestCtx
type FileUploadCtx = model.FileUploadCtx
type User = model.User
type WebsocketCtx = model.WebsocketCtx
type TreeRoot = model.TreeRoot
type TreeCallCtx = model.TreeCallCtx
type BaseBranch = model.BaseBranch
type DocItem = model.DocItem
type WebsocketOptions = model.WebsocketOptions

var Router = model.Router

func Togo(cobject *C.PyObject) *PyObject {
	return (*PyObject)(cobject)
}

func Toc(object *PyObject) *C.PyObject {
	return (*C.PyObject)(object)
}

// 2019-11-16T14:20:08+00:00
// b2s and s2b are borrowed from fasthttp
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func s2b(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

//Iterator's Next
func PyIter_Next(o *PyObject) *PyObject {
	return togo(C._go_PyIter_Next(toc(o)))
}

//Generator Type
//C.PyGen_Type is defined in iap_patched.h as "_go_PyGen_Type"
var Generator = togo((*C.PyObject)(unsafe.Pointer(&C.PyGen_Type)))

func PyGen_Check(o *PyObject) bool {
	return C._go_PyGen_Check(toc(o)) != 0
}

func PyGen_CheckExact(o *PyObject) bool {
	return C._go_PyGen_CheckExact(toc(o)) != 0
}

func PyTraceBack_Print(v *C.PyObject, f *PyObject) int {
	i := C.PyTraceBack_Print(v, toc(f))
	return int(i)
}

/*
Eample: Usage

-- py routine,returns a generator --
def example(*argv,**kw):
    for i in range(10):
        yield "(Ex %s again %s %s)" % (i,argv, kw)

//(a PyGenObject by Call python function in golang)
example :=
next := python3.PyBuiltin_Get("next")
argv := python3.PyTuple_New(1)
python3.PyTuple_SetItem(argv, 0, example)
for {
    m := next.CallObject(argv2)
    if m == nil {
        break
    }
    fmt.Printf("m=%v\n", python3.PyUnicode_AsUTF8(m))
}
*/
func PyBuiltin_Get(name string) *PyObject {
	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))
	return togo(C._go_PyBuiltin_Get(c_name))
}

// PythonError is the type used to represent exceptions that happen in Python
// code.
type PythonError struct {
	Exception string
	Traceback string
}

func (e PythonError) Error() string {
	// Python itself prints the exception after the
	// traceback, so we follow suit.
	return "Python exception: " + e.Traceback + e.Exception
}

// Borrowed From wsgi.go (Original name: pythonError)
// PythonError returns the current Python exception as an error
// value, clearing the exception state in the process.
func GetPythonError() error {
	var exc, val, tb *C.PyObject
	var err PythonError
	C.PyErr_Fetch(&exc, &val, &tb)
	C.PyErr_NormalizeException(&exc, &val, &tb)
	if exc == nil {
		return nil
	}
	defer C.Py_DecRef(exc)
	defer C.Py_DecRef(val)
	if tb != nil {
		defer C.Py_DecRef(tb)
	}
	excs := C.PyObject_Str(val)
	if excs == nil {
		C.PyErr_Clear()
		return PythonError{"Double exception, bailing!", "No traceback, sorry.\n"}
	}
	defer C.Py_DecRef(excs)
	err.Exception = PyUnicode_AsUTF8(togo(excs))
	if tb == nil {
		err.Traceback = "(traceback is empty)\n"
		return err
	}
	iomodule := PyImport_ImportModule("io")
	sioConstructor := iomodule.GetAttrString("StringIO")
	emptyTuple := PyTuple_New(0)
	sio := sioConstructor.CallObject(emptyTuple)
	PyTraceBack_Print(tb, sio)
	getvalue := sio.GetAttrString("getvalue")
	value := getvalue.Call(emptyTuple, Py_None)
	if value == nil {
		err.Traceback = "(empty)"
	} else {
		err.Traceback = PyUnicode_AsUTF8(value)
	}
	return err
}

//
// Python module to pass to pythou routines for handling request.
//

var ObjshCModule *PyObject  //Objsh module in iap_patched.c
var ObjshPyModule *PyObject //Objsh module in iap_pathed.go
var ctxObjectCreator *PyObject
var ctxWebsocketObjectCreator *PyObject
var ctxFileUploadObjectCreator *PyObject
var treeCallCtxObjectCreator *PyObject

//export goUserGetter
func goUserGetter(ptr unsafe.Pointer, cname *C.char) *C.PyObject {
	if user, ok := (pointer.Restore(ptr)).(User); ok {
		switch name := C.GoString(cname); name {
		case "Username":
			return toc(PyUnicode_FromString(user.Username()))
		}
	}
	return toc(PyUnicode_FromString("guest"))
}

// Load cfastjob module
func GetObjshModule() *PyObject {
	if ObjshCModule == nil {
		ObjshCModule = togo(C.PyInit_objshModule())
	}
	return ObjshCModule
}
// load fastjob module
func GetObjshPyModule() *PyObject {
    // 2019-11-21T02:01:19+00:00
    // Objsh is an old name, will be changed to fastjob graduately.
	if ObjshPyModule == nil {
		ObjshPyModule = loadPyModuleFastjob()
	}
	return ObjshPyModule
}

// CtxFamily is a common interface for RequextCtx and FileUploadCtx
// This enable a "ctx" could be either one which makes shorter code
// in goRemoteAddr(), goCtxWrite(), goCtxPeek()
type CtxFamily interface {
	Peek(string) []byte
	Write([]byte) (int, error)
	RemoteAddr() net.Addr
}

//export goRemoteAddr
func goRemoteAddr(ptr unsafe.Pointer) *C.PyObject {
	/*
	   // required? the next 4 lines
	   // maybe not
	   runtime.LockOSThread()
	   gil := PyGILState_Ensure()
	   defer PyGILState_Release(gil)
	   defer runtime.UnlockOSThread()
	*/

	var ctx CtxFamily
	var ok bool
	ctx, ok = (pointer.Restore(ptr)).(*RequestCtx)
	if !ok {
		ctx, ok = (pointer.Restore(ptr)).(*FileUploadCtx)
		if !ok {
			return toc(Py_None)
		}
	}
	addr := fmt.Sprintf("%v", ctx.RemoteAddr())
	return toc(PyUnicode_FromString(addr))
	/*
	       if ctx, ok := (pointer.Restore(ptr)).(*RequestCtx); ok {
	           addr := fmt.Sprintf("%v", ctx.Ctx.RemoteAddr())
	           return toc(PyUnicode_FromString(addr))
	       }
	   return toc(Py_None)
	*/
}

//export goCtxWrite
func goCtxWrite(ptr unsafe.Pointer, argPyBytes *C.PyObject) int {
	/*
	   // required? the next 4 lines
	   // maybe not
	   runtime.LockOSThread()
	   gil := PyGILState_Ensure()
	   defer PyGILState_Release(gil)
	   defer runtime.UnlockOSThread()
	*/

	var ctx CtxFamily
	var ok bool
	ctx, ok = (pointer.Restore(ptr)).(*RequestCtx)
	if !ok {
		ctx, ok = (pointer.Restore(ptr)).(*FileUploadCtx)
		if !ok {
			return 0
		}
	}
	size := C.PyBytes_Size(argPyBytes)
	cBytes := C.PyBytes_AsString(argPyBytes)
	var goBytes []byte
	goBytes = (*[1 << 30]byte)(unsafe.Pointer(cBytes))[0:size]
	bytesent, _ := ctx.Write(goBytes)
	return bytesent

	/*
	   if ctx, ok := (pointer.Restore(ptr)).(*RequestCtx); ok {

	       size := C.PyBytes_Size(argPyBytes)
	       cBytes := C.PyBytes_AsString(argPyBytes)
	       var goBytes []byte
	       goBytes = (*[1 << 30]byte)(unsafe.Pointer(cBytes))[0:size]
	       ctx.Ctx.Write(goBytes)

	   }
	*/
}

//export goCtxPeek
func goCtxPeek(ptr unsafe.Pointer, argPyUnicode *C.PyObject) *C.PyObject {
	var ctx CtxFamily
	var ok bool
	ctx, ok = (pointer.Restore(ptr)).(*RequestCtx)
	if !ok {
		ctx, ok = (pointer.Restore(ptr)).(*FileUploadCtx)
		if !ok {
			return toc(Py_None)
		}
	}
	key := PyUnicode_AsUTF8(togo(argPyUnicode))
    value := ctx.Peek(key)
    if len(value) == 0{
        // key for peeking is not in query string
        return toc(Py_None)
    }else{
        return toc(PyUnicode_FromString(b2s(value)))
    }
        
	/*
	   if ctx, ok := (pointer.Restore(ptr)).(*RequestCtx); ok {
	       key := PyUnicode_AsUTF8(togo(argPyUnicode))
	       value := ctx.Args.Peek(key)
	       return toc(PyUnicode_FromString(b2s(value)))
	   } else if ctx, ok := (pointer.Restore(ptr)).(*FileUploadCtx); ok {
	       key := PyUnicode_AsUTF8(togo(argPyUnicode))
	       value := ctx.Args.Peek(key)
	       return toc(PyUnicode_FromString(b2s(value)))
	   }
	   return toc(Py_None)
	*/
}

//export goCtxSendfile
func goCtxSendfile(ptr unsafe.Pointer, argPyUnicode *C.PyObject) {
	if ctx, ok := (pointer.Restore(ptr)).(*RequestCtx); ok {
		/*
		   // required? the next 4 lines
		   // maybe not
		   runtime.LockOSThread()
		   gil := PyGILState_Ensure()
		   defer PyGILState_Release(gil)
		   defer runtime.UnlockOSThread()
		*/
		path := PyUnicode_AsUTF8(togo(argPyUnicode))
		fmt.Println("sendfile ", path)
		ctx.Ctx.SendFile(path)
	}
}

//export goCtxRedirect
func goCtxRedirect(ptr unsafe.Pointer, argPyUnicode *C.PyObject, status C.int) {
	if ctx, ok := (pointer.Restore(ptr)).(*RequestCtx); ok {
		/*
		   // required? the next 4 lines
		   // maybe not
		   runtime.LockOSThread()
		   gil := PyGILState_Ensure()
		   defer PyGILState_Release(gil)
		   defer runtime.UnlockOSThread()
		*/
		path := PyUnicode_AsUTF8(togo(argPyUnicode))
		ctx.Ctx.Redirect(path, int(status))
	}
}

func NewPyCtxObject(metadata *PyObject, ctx *RequestCtx) *PyObject {

	runtime.LockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	defer runtime.UnlockOSThread()

	argv := PyTuple_New(1)
	PyTuple_SetItem(argv, 0, metadata)
	pyCtxInst := ctxObjectCreator.Call(argv, Py_None)
	ctxptr := pointer.Save(ctx)
	reqptr := pointer.Save(ctx.Ctx.Request)
	var reqheaderptr unsafe.Pointer
	reqheaderptr = pointer.Save(&ctx.Ctx.Request.Header)
	respptr := pointer.Save(ctx.Ctx.Response)
	respheaderptr := pointer.Save(&ctx.Ctx.Response.Header)
	var userptr unsafe.Pointer
	if ctx.User == nil {
		//PublicMode
		userptr = nil
	} else {
		//ProtectMode
		userptr = pointer.Save(ctx.User)
	}
	C.SetCtxPtr(toc(pyCtxInst), ctxptr, reqptr, reqheaderptr, respptr, respheaderptr, userptr)
	return pyCtxInst
}


//export goObjshRouterGet
func goObjshRouterGet(path *C.PyObject, handler *C.PyObject, acl int) {
	urlpath := PyUnicode_AsUTF8(togo(path))
    Get(urlpath, togo(handler), acl, false)
}

//export goObjshRouterPost
func goObjshRouterPost(path *C.PyObject, handler *C.PyObject, acl int) {
	urlpath := PyUnicode_AsUTF8(togo(path))
    Get(urlpath, togo(handler), acl, true)
}

//export goObjshRouterWebsocket
func goObjshRouterWebsocket(path *C.PyObject, handler *C.PyObject, acl int) {
	urlpath := PyUnicode_AsUTF8(togo(path))
	Websocket(urlpath, togo(handler), acl)
}

//export goObjshRouterFileUpload
func goObjshRouterFileUpload(path *C.PyObject, handler *C.PyObject, acl int) {
	//Router := objsh.Router
	urlpath := PyUnicode_AsUTF8(togo(path))
	FileUpload(urlpath, togo(handler), acl, false)
}

//export goCallFunc
func goCallFunc(typename *C.char, ptr unsafe.Pointer, funcname *C.char, args *C.PyObject) *C.PyObject {

	//runtime.LockOSThread()
	//gil := PyGILState_Ensure()
	//defer PyGILState_Release(gil)
	//defer runtime.UnlockOSThread()

	tname := C.GoString(typename)
	fname := C.GoString(funcname)
	var ret []reflect.Value
	switch tname {
	case "ResponseHeader":
		if header, ok := (pointer.Restore(ptr)).(*fasthttp.ResponseHeader); ok {
			v := reflect.ValueOf(header)
			m := v.MethodByName(fname)
			if !m.IsValid() {
				return toc(Py_None)
			}
			goArgs := togo(args)
			size := PyTuple_Size(goArgs)
			param := make([]reflect.Value, size)
			// 這裡有個問題是，假設所有傳入參數都是字串(content-length不是，要在python中先轉成字串)
			for i := 0; i < size; i++ {
				arg := PyTuple_GetItem(goArgs, i)
				param[i] = reflect.ValueOf(PyUnicode_AsUTF8(arg))
			}
			ret = m.Call(param) //ret is []reflect.Value
			/*
			   if len(ret) == 0 {
			       return toc(Py_None)
			   }
			   // 這裡有個問題是，假設只有一個回傳參數，而且是字串
			   pyobj := PyUnicode_FromString(ret[0].Interface().(string))
			   return toc(pyobj)
			*/
		}
	case "RequestHeader":
		if header, ok := (pointer.Restore(ptr)).(*fasthttp.RequestHeader); ok {
			v := reflect.ValueOf(header)
			m := v.MethodByName(fname)
			//fmt.Println("call RequestHeader ", fname, header.Peek("title"))
			if !m.IsValid() {
				return toc(Py_None)
			}
			goArgs := togo(args)
			size := PyTuple_Size(goArgs)
			param := make([]reflect.Value, size)
			// 這裡有個問題是，假設所有傳入參數都是字串
			for i := 0; i < size; i++ {
				arg := PyTuple_GetItem(goArgs, i)
				param[i] = reflect.ValueOf(PyUnicode_AsUTF8(arg))
			}
			ret = m.Call(param) //ret is []reflect.Value
		}
	}

	//return result
	if ret == nil {
		//No result
		return toc(Py_None)
	}

	length := len(ret)
	if length == 0 {
		return toc(Py_None)
	}
	// 這裡有個問題是，假設只有一個回傳參數，而且是字串
	var tuple *PyObject
	if length > 1 {
		tuple = PyTuple_New(length)
	}
	for i, v := range ret {
		var pyvalue *PyObject
		switch v.Interface().(type) {
		case string:
			pyvalue = PyUnicode_FromString(v.String())
		case int, int8, int16, int32, int64: //reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			pyvalue = PyLong_FromGoInt64(v.Int())
		case []byte:
			pyvalue = PyUnicode_FromString(string(v.Bytes()))
		default:
			vv := reflect.ValueOf(v)
			pyvalue = PyUnicode_FromString(fmt.Sprintf("unhandled type '(Type:%s)' of %#v", v.Type(), vv.Type()))
		}

		if length == 1 {
			return toc(pyvalue)
		} else {
			PyTuple_SetItem(tuple, i, pyvalue)
		}
	}
	return toc(tuple)
}
func GenRequestCtx4Python(ctx *RequestCtx) *PyObject {
	//objsh.RequestCtx is model.RequestCtx defined in userauth.go
	ret := PyDict_New()
	/*
	   user := PyDict_New()
	   if ctx.User != nil {
	       //fmt.Fprint(ctx, ctx.User.Username())
	       PyDict_SetItem(user, PyUnicode_FromString("Username"), PyUnicode_FromString(ctx.User.Username()))
	   } else {
	       PyDict_SetItem(user, PyUnicode_FromString("Username"), PyUnicode_FromString("guest"))
	   }
	   PyDict_SetItem(ret, PyUnicode_FromString("user"), user)
	*/

	kw := PyDict_New()
	args := ctx.Args
	args.VisitAll(func(k []byte, v []byte) {
		PyDict_SetItem(kw, PyUnicode_FromString(string(k)), PyUnicode_FromString(string(v)))
	})
	PyDict_SetItem(ret, PyUnicode_FromString("kw"), kw)

	reqHeader := PyDict_New()
	headers := [...]string{"Cookie", "User-Agent", "Host", "Referer", "Accept-Language"}
	for _, header := range headers {
		PyDict_SetItem(reqHeader, PyUnicode_FromString(header), PyUnicode_FromString(string(ctx.Ctx.Request.Header.Peek(header))))
	}
	PyDict_SetItem(ret, PyUnicode_FromString("header"), reqHeader)

	C.Py_IncRef(toc(ret))
	return ret
}

//register request hander of Get and Post
func Get(urlpath string, handler *PyObject, acl int, post bool){
    
	gohandler := handler
	ret := func(ctx *RequestCtx) {

		//f.Call need GIL miso
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		pyCtxMetadata := GenRequestCtx4Python(ctx)
		pyCtxInst := NewPyCtxObject(pyCtxMetadata, ctx) 

		//call python request handler
		argv := PyTuple_New(1)
		PyTuple_SetItem(argv, 0, pyCtxInst)
		ret := gohandler.CallObject(argv) //, Py_None)
		if ret == nil {
			err := GetPythonError()
			fmt.Fprintf(ctx.Ctx, "Failed to call %v, reason:%v", urlpath, err)
		} else if PyUnicode_Check(ret) {
			fmt.Fprint(ctx.Ctx, PyUnicode_AsUTF8(ret))
		} else if PyBytes_Check(ret) {
			size := C.int(PyBytes_Size(ret))
			cBytes := C.PyBytes_AsString(toc(ret))
			var goBytes []byte
			goBytes = (*[1 << 30]byte)(unsafe.Pointer(cBytes))[0:size]
			ctx.Ctx.Write(goBytes)
		} else if ret == Py_None {
			// nothing
		} else {
			log.Println("Warning: only unicode or bytes can be returned from python")
        }
        // 不要送到defer去 release,　會有memory access的問題
		PyGILState_Release(gil)
		runtime.UnlockOSThread()
        
	}
	if post {
		Router.Post(urlpath, ret, acl)
	} else {
		Router.Get(urlpath, ret, acl)
    }
    fmt.Println("Regiter Get ",urlpath)
}

/*
Wsgi ---- Not implemented yet ----
*/
func genWsgiEnviron(ctx *RequestCtx) {

}

//register request hander of WSGI(Not implemented)
func Wsgi(urlpath string, handler *PyObject, acl int) {
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)

	var ret model.RequestHandler
	//gohandler := handler
	ret = func(ctx *RequestCtx) {

		//f.Call need GIL
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		/*
		   pyEnviron := GenWsgiEnviron(ctx)
		   pyCtxInst := NewPyCtxObject(pyCtxMetadata, ctx)

		   //call python request handler
		   argv := PyTuple_New(1)
		   PyTuple_SetItem(argv, 0, pyCtxInst)
		   ret := gohandler.Call(argv, Py_None)
		   fmt.Fprint(ctx.Ctx, PyUnicode_AsUTF8(ret))
		   //fmt.Println("run", urlpath, "got", PyUnicode_AsUTF8(ret))
		*/
		PyGILState_Release(gil)
		runtime.UnlockOSThread()
    }
	Router.Get(urlpath, ret, acl)
}

/*
 Websocket
*/

func NewPyWebsocketCtxObject(metadata *PyObject, ctx *WebsocketCtx) *PyObject {

	runtime.LockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	defer runtime.UnlockOSThread()

	argv := PyTuple_New(1)
	PyTuple_SetItem(argv, 0, metadata)
	pyCtxInst := ctxWebsocketObjectCreator.Call(argv, Py_None)
	//C.SetWebsocketCtxPtr(toc(pyCtxInst), pointer.Save(goObj))

	if ctx.User == nil {
		C.SetWebsocketCtxPtr(toc(pyCtxInst), pointer.Save(ctx), nil)
	} else {
		C.SetWebsocketCtxPtr(toc(pyCtxInst), pointer.Save(ctx), pointer.Save(ctx.User))
	}

	return pyCtxInst
}

/*
func PyWebsocketCtxObjectOnMessage(pyWsCtxObj *PyObject, mesg string) {

    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    gil := PyGILState_Ensure()
    defer PyGILState_Release(gil)

    C.WebsocketCtxOnMessage(toc(pyWsCtxObj), toc(PyUnicode_FromString(mesg)))

}
*/

//export goWebsocketCtxSend
func goWebsocketCtxSend(ptr unsafe.Pointer, mesg *C.PyObject) int {
	if pyWsCtx, ok := (pointer.Restore(ptr)).(*WebsocketCtx); ok {
		//runtime.LockOSThread()
		//gil := PyGILState_Ensure()
		//defer PyGILState_Release(gil)
		pyWsCtx.Send(PyUnicode_AsUTF8(togo(mesg)))
		return 1
	}
	return 0
}

//export goGetUserData
func goGetUserData(ptr unsafe.Pointer, cargs *C.PyObject) *C.PyObject {
	if user, ok := (pointer.Restore(ptr)).(model.User); ok {
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		defer PyGILState_Release(gil)
		defer runtime.UnlockOSThread()

		args := togo(cargs)
		key := PyTuple_GetItem(args, 0)
		if value, ok := user.GetMetadata(PyUnicode_AsUTF8(key)); ok {
			str := PyUnicode_FromString(value)
			return toc(str)
		}
	}
	return toc(Py_None)
}

//export goUserMetadata
func goUserMetadata(ptr unsafe.Pointer) *C.PyObject {
	ret := PyDict_New()
	if user, ok := (pointer.Restore(ptr)).(model.User); ok {
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		defer PyGILState_Release(gil)
		defer runtime.UnlockOSThread()
		metadata := user.Metadata()

		if metadata != nil {
			for k, v := range metadata {
				//ret[k] = v
				PyDict_SetItem(ret, PyUnicode_FromString(k), PyUnicode_FromString(v))
			}
		}
	}
	return toc(ret)
}

//export goWsCtxAddEventListener
func goWsCtxAddEventListener(ptr unsafe.Pointer, evtname *C.char, token4remove *C.char, handler *C.PyObject) int {
	if wsCtx, ok := (pointer.Restore(ptr)).(*WebsocketCtx); ok {
		name := C.GoString(evtname)
		switch name {
		case "Message":
			wsCtx.On("Message", C.GoString(token4remove), func(mesg string) {
				runtime.LockOSThread()
				gil := PyGILState_Ensure()
				defer PyGILState_Release(gil)
				defer runtime.UnlockOSThread()
				args := PyTuple_New(1)
				PyTuple_SetItem(args, 0, PyUnicode_FromString(mesg))
				togo(handler).CallObject(args)
			})
		case "Close":
			wsCtx.On("Close", C.GoString(token4remove), func() {
				runtime.LockOSThread()
				gil := PyGILState_Ensure()
				defer PyGILState_Release(gil)
				defer runtime.UnlockOSThread()
				args := PyTuple_New(0)
				togo(handler).CallObject(args)
			})
		}
		return 1
	}
	return 0
}

func loadPyModuleFastjob() *PyObject {
	//content is copied from src/fastjob/python/fastjob.py
	content := `
# 注意：這段PY有錯誤會在compile時報錯為： segmentation violatoin
# will be copied to iap_patched.go
import traceback

#
# Routing
#

# 2019-11-21T02:13:58+00:00
#   Will be deprecated, use "acl" instead
PublicMode = 1
TraceMode = 2
ProtectMode = 3

class ACL:
    PublicMode = 1
    TraceMode = 2
    ProtectMode = 3

# cObjshRouter is an ObjshRouter instance to objsh.Router(golang). 
# (cObjshRouter is defined in iap_patched.c)
cObjshRouter = None



import importlib, sys
class ReloadableRouterWrapper(object):
    def __init__(self):
        self.handlers = {}

    def reloadModule(self,name):
        try:
            importlib.reload(sys.modules[name])
            return 'reload module "' + name +'" completed' 
        except:
            return traceback.format_exc()

    def register(self,method,path,acl):
        def f(handler):
            try:
                self.handlers[path]
            except KeyError:
                # register this path at frist time
                def registeredHandler(*args,**kw):
                    return self.handlers[path](*args,**kw)
                if method == 'Get':
                    cObjshRouter.Get(path,registeredHandler,acl)
                elif method == 'Post':
                    cObjshRouter.Post(path,registeredHandler,acl)
                elif method == 'Websocket':
                    cObjshRouter.Websocket(path,registeredHandler,acl)
                elif method == 'FileUpload':
                    cObjshRouter.FileUpload(path,registeredHandler,acl)
                else:
                    raise NotImplementedError(method+' not implmented')
            self.handlers[path] = handler
        return f

class RouterWrapper(object):
    def __init__(self):
        self.reloadableRouter = ReloadableRouterWrapper()

    def reloadModule(self,name):
        return  self.reloadableRouter.reloadModule(name)

    def Get(self,path,acl,reloadable=False):	
        if reloadable:
            return self.reloadableRouter.register('Get',path,acl)
        else:
            def f(handler):
                cObjshRouter.Get(path,handler,acl)
            return f

    def Post(self,path,acl,reloadable=False):	
        if reloadable:
            return self.reloadableRouter.register('Post',path,acl)
        else:
            def f(handler):
                cObjshRouter.Post(path,handler,acl)
            return f

    def Websocket(self,path,acl,reloadable=False):
        if reloadable:
            return self.reloadableRouter.register('Websocket',path,acl)
        else:
            def f(handler):
                cObjshRouter.Websocket(path,handler,acl)
            return f

    def FileUpload(self,path,acl,reloadable=False):
        if reloadable:
            return self.reloadableRouter.register('FileUpload',path,acl)
        else:
            def f(handler):
                cObjshRouter.FileUpload(path,handler,acl)
            return f
    # Not implemented yet
    #def Wsgi(self,path,acl,reloadable=False):	
    #	def f(handler):
    #		cObjshRouter.Wsgi(path,handler,acl)
    #	return f

Router = RouterWrapper()
#ReloadableRouter = ReloadableRouterWrapper()

#
# Tree API
#

# cObjshTree is an ObjshTree instance to objsh.Tree(golang). 
# (cObjshTree is defined in iap_patched.c)
cObjshTree = None

class BaseBranch(object):
    def __init__(self,name=None):
        self.name = name
        self.exportableNames = []
        self.exportableDocs = {}
    def getExportableNames(self):
        return self.exportableNames[:]
    #def beReady(self,tree):
    #    # return False if root.SureReady is going to be called later manually
    #    return True
    def beReady(self,tree):
        raise NotImplementError('%s.beReady not implemented' % self)
    def _beReady(self,treeName):
        try:
            return self.beReady(getattr(Tree,treeName))
        except:
            traceback.print_exc()
            raise
    def __call__(self,methodName,ctx):
        if not methodName in self.exportableNames:
            raise AttributeError(methodName + ' is not exported')
        try:
            getattr(self,methodName)(ctx)
        except:
            ctx.reject(400,traceback.format_exc()) 
    def export(self,*funcs):
        self.exportableNames = []
        self.exportableDocs = {}
        for func in funcs:
            print("exporting",func.__name__)
            self.exportableNames.append(func.__name__)
            self.exportableDocs[func.__name__] = func.__doc__

class PesudoTree(object):
    def __init__(self,name):
        self.name = name
    def addBranch(self,branchObj,branchName=None):
        try:
            if branchName is not None:
                assert isinstance(branchName,str), 'addBranch(branchObj,branchName) where branchName should be string'
                branchObj.name = branchName
            cObjshTree.AddBranch(branchObj,self.name)
        except:
            print(traceback.format_exc())
    addBranchWithName = addBranch
    def sureReady(self,branchObj):
        cObjshTree.SureReady(self.name, branchObj.name)

class PyTreeWrapper(object):
    # Wrap trees in golang for python scripts
    def __init__(self):
        pass

    def addTree(self,name,*args):
        # Called in golang, to make a python statement be valid, such as
        # Tree.UnitTest.addBranch, Tree.Member.addBranch
        tree = PesudoTree(name)
        setattr(self,name,tree)


initCallables = []
def callWhenRunning(func,*args,**kw):
    initCallables.append((func,args,kw))
def callInitCallables():
    print('Call initial callables' * 20,initCallables)
    for func,args,kw in initCallables:
        func(*args,**kw)

# Prefer GoTrees than Tree
GoTrees = Tree = PyTreeWrapper()
# Utility var to be called in golang
_addTree = GoTrees.addTree

__all__ = ['Router','ACL','GoTrees','Tree','BaseBranch','callWhenRunning',
           'PublicMode','TraceMode','ProtectMode']
`

	// Let make "from fastjob impot *" happen
    // here the "fastjob" make sense to user's script (such as in handlers.py)

    compile := PyBuiltin_Get("compile")
	args := PyTuple_New(3)
	PyTuple_SetItem(args, 0, PyUnicode_FromString(content))
	PyTuple_SetItem(args, 1, PyUnicode_FromString("<string>"))
	PyTuple_SetItem(args, 2, PyUnicode_FromString("exec"))
	code := compile.CallObject(args)
    
	if code == nil {
		err := GetPythonError()
		fmt.Errorf("failed to compile RouterWrapper:\n%v\n", err)
	}
	c_fastjob := C.CString("fastjob")
	defer C.free(unsafe.Pointer(c_fastjob))
	m := C.PyImport_ExecCodeModule(c_fastjob, toc(code))

	if m == nil {
		err := GetPythonError()
		fmt.Errorf("failed to load fastjob module:\n%v\n", err)
	}
	return togo(m)
}

func GenRequestWsCtx4Python(wsCtx *model.WebsocketCtx) *PyObject {
	//objsh.RequestCtx is model.RequestCtx defined in userauth.go
	ret := PyDict_New()
	user := PyDict_New()
	if wsCtx.User != nil {
		//fmt.Fprint(ctx, ctx.User.Username())
		PyDict_SetItem(user, PyUnicode_FromString("Username"), PyUnicode_FromString(wsCtx.User.Username()))
	} else {
		PyDict_SetItem(user, PyUnicode_FromString("Username"), PyUnicode_FromString("guest"))
	}
	PyDict_SetItem(ret, PyUnicode_FromString("User"), user)

	kw := PyDict_New()
	args := wsCtx.Args
	args.VisitAll(func(k []byte, v []byte) {
		PyDict_SetItem(kw, PyUnicode_FromString(string(k)), PyUnicode_FromString(string(v)))
	})
	PyDict_SetItem(ret, PyUnicode_FromString("kw"), kw)

	return ret
}

//export goPrint
func goPrint(mesg *C.char) {
	fmt.Println(C.GoString(mesg))
}
func Websocket(urlpath string, handler *PyObject, acl int) {
    WebsocketWithOptions(urlpath, handler, acl, nil) 
}
func WebsocketWithOptions(urlpath string, handler *PyObject, acl int, options *WebsocketOptions) {
	//func Websocket(urlpath string, pypath string, acl int) {
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	Router.WebsocketWithOptions(urlpath, func(wsCtx *WebsocketCtx) {

		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		defer PyGILState_Release(gil)
		defer runtime.UnlockOSThread()

		pyCtxMetadata := GenRequestWsCtx4Python(wsCtx)
		pyCtxInst := NewPyWebsocketCtxObject(pyCtxMetadata, wsCtx) //pyCtxWrapper)

		//call python request handler
		argv := PyTuple_New(1)
		PyTuple_SetItem(argv, 0, pyCtxInst)
		handler.Call(argv, Py_None)

	}, acl,options)
}

/*
 FileUpload
*/

func NewPyFileUploadCtxObject(metadata *PyObject, ctx *FileUploadCtx) *PyObject {
	runtime.LockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	defer runtime.UnlockOSThread()

	argv := PyTuple_New(1)
	PyTuple_SetItem(argv, 0, metadata)
	pyCtxInst := ctxFileUploadObjectCreator.Call(argv, Py_None)
	/*
	   if ctx.User == nil {
	       C.SetFileUploadCtxPtr(toc(pyCtxInst), pointer.Save(ctx), nil)
	   } else {
	       C.SetFileUploadCtxPtr(toc(pyCtxInst), pointer.Save(ctx), pointer.Save(ctx.User))
	   }
	*/

	ctxptr := pointer.Save(ctx)
	reqptr := pointer.Save(ctx.Ctx.Request)
	var reqheaderptr unsafe.Pointer
	reqheaderptr = pointer.Save(&ctx.Ctx.Request.Header)
	respptr := pointer.Save(ctx.Ctx.Response)
	respheaderptr := pointer.Save(&ctx.Ctx.Response.Header)
	var userptr unsafe.Pointer
	if ctx.User == nil {
		//PublicMode
		userptr = nil
	} else {
		//ProtectMode
		userptr = pointer.Save(ctx.User)
	}
	C.SetFileUploadCtxPtr(toc(pyCtxInst), ctxptr, reqptr, reqheaderptr, respptr, respheaderptr, userptr)

	return pyCtxInst
}

func GenPyMetadata4FileUpload(ctx *FileUploadCtx) *PyObject {
	//objsh.RequestCtx is model.RequestCtx defined in userauth.go
	ret := PyDict_New()
	user := PyDict_New()
	if ctx.User != nil {
		//fmt.Fprint(ctx, ctx.User.Username())
		PyDict_SetItem(user, PyUnicode_FromString("Username"), PyUnicode_FromString(ctx.User.Username()))
	} else {
		PyDict_SetItem(user, PyUnicode_FromString("Username"), PyUnicode_FromString("guest"))
	}
	PyDict_SetItem(ret, PyUnicode_FromString("User"), user)

	kw := PyDict_New()
	args := ctx.Args
	args.VisitAll(func(k []byte, v []byte) {
		PyDict_SetItem(kw, PyUnicode_FromString(string(k)), PyUnicode_FromString(string(v)))
	})
	PyDict_SetItem(ret, PyUnicode_FromString("kw"), kw)

	C.Py_IncRef(toc(ret))
	return ret
}

// FileUpload wrapper Python ctx.FileUpload to golang ctx.FileUpload
func FileUpload(urlpath string, handler *PyObject, acl int, post bool) {
	//gil := PyGILState_Ensure()
	//defer PyGILState_Release(gil)

	gohandler := handler
	Router.FileUpload(urlpath, func(ctx *FileUploadCtx) {
		//need GIL
		runtime.LockOSThread()
		gil := PyGILState_Ensure()

		pyCtxMetadata := GenPyMetadata4FileUpload(ctx)
		pyCtxInst := NewPyFileUploadCtxObject(pyCtxMetadata, ctx)
		//call python request handler
		argv := PyTuple_New(1)
		PyTuple_SetItem(argv, 0, pyCtxInst)
		ret := gohandler.CallObject(argv) //, Py_None)
		if ret == nil {
			err := GetPythonError()
			fmt.Fprintf(ctx.Ctx, "Failed to call %v, reason:%v", urlpath, err)
		} else if PyUnicode_Check(ret) {
			fmt.Fprint(ctx.Ctx, PyUnicode_AsUTF8(ret))
		} else if PyBytes_Check(ret) {
			size := C.int(PyBytes_Size(ret))
			cBytes := C.PyBytes_AsString(toc(ret))
			var goBytes []byte
			goBytes = (*[1 << 30]byte)(unsafe.Pointer(cBytes))[0:size]
			ctx.Ctx.Write(goBytes)
		} else if ret == Py_None {
			// nothing
		} else {
			log.Println("Warning: only unicode or bytes can be returned from python")
        }
		PyGILState_Release(gil)
		runtime.UnlockOSThread()
	}, acl)
}

//export goFileUploadCtxGetter
func goFileUploadCtxGetter(ptr unsafe.Pointer, cname *C.char) *C.PyObject {
	if fuCtx, ok := (pointer.Restore(ptr)).(*FileUploadCtx); ok {
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		defer PyGILState_Release(gil)
		defer runtime.UnlockOSThread()
		name := C.GoString(cname)
		switch name {
		case "Filename":
			return toc(PyUnicode_FromString(fuCtx.Filename))
		case "Filesize":
			return toc(PyLong_FromGoInt64(fuCtx.Filesize))
		}
	}
	return toc(Py_None)
}

//export goFileUploadCtxSaveTo
func goFileUploadCtxSaveTo(ptr unsafe.Pointer, cpath *C.PyObject) int {
	if fuCtx, ok := (pointer.Restore(ptr)).(*FileUploadCtx); ok {
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		defer PyGILState_Release(gil)
		defer runtime.UnlockOSThread()

		path := PyUnicode_AsUTF8(togo(cpath))
		if err := fuCtx.SaveTo(path); err != nil {
			log.Printf("FileUploadCtx SaveTo(%v) error:%v", path, err)
			return -1
		}
	}
	return 0
}

/*
 Init module
*/
func InitIapPatchedModules() {

	// 載入製作ctx的類別與constructor(defined in iap_patched.c)
	// 這個 objshModule是定義在c裡面的py module
	objshModule := GetObjshModule()
	if objshModule == nil {
		panic("objshModule is nil")
	}
	ctxObjectCreator = objshModule.GetAttrString("CtxObject")
	if ctxObjectCreator == nil {
		panic("ctxObjectCreator is nil")
	}
	ctxWebsocketObjectCreator = objshModule.GetAttrString("WebsocketCtxObject")
	if ctxWebsocketObjectCreator == nil {
		panic("ctxWebsocketObjectCreator is nil")
	}
	ctxFileUploadObjectCreator = objshModule.GetAttrString("FileUploadCtxObject")
	if ctxFileUploadObjectCreator == nil {
		panic("ctxFileUploadObjectCreator is nil")
	}
	treeCallCtxObjectCreator = objshModule.GetAttrString("TreeCallCtxObject")
	if treeCallCtxObjectCreator == nil {
		panic("treeCallCtxObjectCreator is nil")
	}

	//這個 m 是上面定義在golang裡面的module
	pyFastjobModule := GetObjshPyModule()

	// Create cObjshRouter for python module "objsh"
	emptyTuple := PyTuple_New(0)
	objshRouterCreator := objshModule.GetAttrString("ObjshRouter")
	objshRouter := objshRouterCreator.Call(emptyTuple, Py_None)
	if objshRouter == nil {
		panic("objshRouter is nil")
	}

	c_cObjshRouter := C.CString("cObjshRouter")
	defer C.free(unsafe.Pointer(c_cObjshRouter))
	C.PyObject_SetAttrString(toc(pyFastjobModule), c_cObjshRouter, toc(objshRouter))

	// Create cObjshRouter for python module "objsh"
	c_cObjshTree := C.CString("cObjshTree")
	defer C.free(unsafe.Pointer(c_cObjshTree))
	objshTreeCreator := objshModule.GetAttrString("ObjshTree")
	objshTree := objshTreeCreator.Call(emptyTuple, Py_None)
	if objshTree == nil {
		panic("objshTree is nil")
	}
	C.PyObject_SetAttrString(toc(pyFastjobModule), c_cObjshTree, toc(objshTree))
}

func CallWhenRunning() {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)

	//這個 m 是上面定義在golang裡面的module
	m := GetObjshPyModule()
	callInitCallables := m.GetAttrString("callInitCallables")
	emptyTuple := PyTuple_New(0)
	callInitCallables.Call(emptyTuple, Py_None)
}

var treeExposedToPython map[string]*TreeRoot

func AddTree(tree *TreeRoot) {
	runtime.LockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	defer runtime.UnlockOSThread()

	addTree := GetObjshPyModule().GetAttrString("_addTree")
	if addTree == nil {
		panic("_addTree() is nil")
	}
	args := PyTuple_New(1)
	PyTuple_SetItem(args, 0, PyUnicode_FromString(tree.Name))
	addTree.CallObject(args)

	// Add to treeExposedToPython for looking up
	if treeExposedToPython == nil {
		treeExposedToPython = make(map[string]*TreeRoot)
	}
	treeExposedToPython[tree.Name] = tree
}

/*
PyBranch implement Branch interface for wrapping Python BaseBranch
*/

func GenPyMetadata4TreeCall(ctx *TreeCallCtx) *PyObject {
	//objsh.RequestCtx is model.RequestCtx defined in userauth.go
	ret := PyDict_New()
	args := PyTuple_New(len(ctx.Args))
	for i := 0; i < len(ctx.Args); i++ {
		PyTuple_SetItem(args, i, PyUnicode_FromString(ctx.Args[i]))
	}
	PyDict_SetItem(ret, PyUnicode_FromString("Args"), args)
	kw := PyDict_New()
	ctx.Kw.VisitAll(func(k []byte, v []byte) {
		PyDict_SetItem(kw, PyUnicode_FromString(string(k)), PyUnicode_FromString(string(v)))
	})
	PyDict_SetItem(ret, PyUnicode_FromString("Kw"), kw)

	C.Py_IncRef(toc(ret))
	return ret
}
func NewPyTreeCallCtxObject(ctx *TreeCallCtx) *PyObject {

	pyCtxMetadata := GenPyMetadata4TreeCall(ctx)

	argv := PyTuple_New(1)
	PyTuple_SetItem(argv, 0, pyCtxMetadata)
	pyCtxInst := treeCallCtxObjectCreator.CallObject(argv)

	ctxptr := pointer.Save(ctx)
	wsctxptr := pointer.Save(ctx.WsCtx)

	var userptr unsafe.Pointer
	user := ctx.WsCtx.GetUser()
	if user == nil {
		//PublicMode
		userptr = nil
	} else {
		//ProtectMode
		userptr = pointer.Save(user)
	}

	C.SetTreeCallCtxPtr(toc(pyCtxInst), ctxptr, wsctxptr, userptr)

	return pyCtxInst
}

//export goTreeCallCtxResolve
func goTreeCallCtxResolve(ptr unsafe.Pointer, jsonstr *C.char) {
	if ctx, ok := (pointer.Restore(ptr)).(*TreeCallCtx); ok {

		//從c.char轉[]byte會比較有效率，目前還不知道怎麼轉，
		//暫時先這樣
		//size := unsafe.Sizeof(unsafe.Pointer(&jsonstr))
		//stdout := C.GoBytes(unsafe.Pointer(&jsonstr), &size)
		stdout := C.GoString(jsonstr)
		result := model.Result{
			Id:      ctx.CmdID,
			Retcode: 0,
			Stdout:  []byte(stdout),
		}
		ctx.DirectResult(&result, true)
	}
}

//export goTreeCallCtxNotify
func goTreeCallCtxNotify(ptr unsafe.Pointer, jsonstr *C.char) {
	if ctx, ok := (pointer.Restore(ptr)).(*TreeCallCtx); ok {
		stdout := C.GoString(jsonstr)
		result := model.Result{
			Id:      ctx.CmdID,
			Retcode: ctx.RetcodeOfNotify,
			Stdout:  []byte(stdout),
		}
		ctx.DirectResult(&result, false)

	}
}

//export goTreeCallCtxReject
func goTreeCallCtxReject(ptr unsafe.Pointer, retcode C.long, jsonstr *C.char) {
	if ctx, ok := (pointer.Restore(ptr)).(*TreeCallCtx); ok {

		//runtime.LockOSThread()
		//gil := PyGILState_Ensure()
		//defer PyGILState_Release(gil)
		//defer runtime.UnlockOSThread()

		//args := togo(cargs)
		retcode := int32(retcode)
		stderr := errors.New(C.GoString(jsonstr))
		//stderr := errors.New("err")

		ctx.Reject(retcode, stderr)
	}
}

/*
func goTreeCallCtxReject(ptr unsafe.Pointer, cargs *C.PyObject) {
    if ctx, ok := (pointer.Restore(ptr)).(*TreeCallCtx); ok {
        runtime.LockOSThread()
        gil := PyGILState_Ensure()
        defer PyGILState_Release(gil)
        defer runtime.UnlockOSThread()

        args := togo(cargs)
        retcode := int32(PyLong_AsLong(PyTuple_GetItem(args, 0)))
        stderr := errors.New(PyUnicode_AsUTF8(PyTuple_GetItem(args, 1)))

        ctx.Reject(retcode, stderr)
    }
    //return toc(Py_None)
}
*/
//export goTreeCallCtxSetBackground
func goTreeCallCtxSetBackground(ptr unsafe.Pointer) *C.PyObject {
	if ctx, ok := (pointer.Restore(ptr)).(*TreeCallCtx); ok {
		/*
		   runtime.LockOSThread()
		   gil := PyGILState_Ensure()
		   defer PyGILState_Release(gil)
		   defer runtime.UnlockOSThread()
		*/

		/*

		   2019-10-22T05:43:09+00:00
		   應該不必那麼複雜，setBackground只有True的方法，沒有取消的必要

		   args := togo(cargs)
		   yes := PyTuple_GetItem(args, 0)
		   if C.PyObject_IsTrue(toc(yes)) == 1 {
		       ctx.SetBackground(true)
		   } else {
		       ctx.SetBackground(false)
		   }
		*/
		ctx.SetBackground(true)
	}
	return toc(Py_None)
}

//PyBranch
type PyBranch struct {
	BaseBranch
	branch *PyObject
}

func (self *PyBranch) BeReady(tree *TreeRoot) {
	
    runtime.LockOSThread()
	gil := PyGILState_Ensure()

	self.InitBaseBranch()
    log.Println("Asking tree <",tree.Name,">'s python branch <",self.Name(), ">to be ready")
    if self.branch  == nil{
		if err := GetPythonError(); err != nil {
			fmt.Println(err)
		}
		panic(fmt.Sprintf("Python self.branch is nil "))
    }    
	beReady := self.branch.GetAttrString("_beReady")
	args := PyTuple_New(1)
	PyTuple_SetItem(args, 0, PyUnicode_FromString(tree.Name))
	ret := beReady.CallObject(args)
	if (ret != nil) && PyBool_Check(ret) {
		yes := C.PyObject_IsTrue(toc(ret))
		if yes == 1 {
			tree.SureReady(self)
		} else {
			fmt.Println("not ready by yes==>", yes)
		}
	} else {
		if err := GetPythonError(); err != nil {
			fmt.Println(err)
		}
		panic(fmt.Sprintf("Python script branch.beReady() should return boolean, but got %v", ret))
	}
	PyGILState_Release(gil)
	runtime.UnlockOSThread()
}
func (self *PyBranch) GetExportableNames(tree *TreeRoot) []string {
	runtime.LockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	defer runtime.UnlockOSThread()

	exportableNames := self.branch.GetAttrString("exportableNames")
	size := PyList_Size(exportableNames)
	var ret []string
	if size > 0 {
		ret = make([]string, size)
		for i := 0; i < size; i++ {
			ret[i] = PyUnicode_AsUTF8(PyList_GetItem(exportableNames, i))
		}
	} else {
		ret = make([]string, 0)
	}

	//Add APIInfo
	exportableDocs := self.branch.GetAttrString("exportableDocs")
	items := PyDict_Items(exportableDocs)
	size = PyList_Size(items)
	if size > 0 {
		now := time.Now()
		ts := now.Unix()
		scrpath := ""
		innerName := ""
		for i := 0; i < size; i++ {
			item := PyList_GetItem(items, i)
			key := PyUnicode_AsUTF8(PyTuple_GetItem(item, 0))
			value := PyUnicode_AsUTF8(PyTuple_GetItem(item, 1))
			apiName := self.Name() + "." + key
			docItem := model.NewDocItem(
				scrpath,
				innerName,
				value,
				ts,
			)
			tree.SetAPIInfo(apiName, &docItem)
		}
	}

	return ret
}

func (self *PyBranch) Call(methodName string, ctx *TreeCallCtx) {

	runtime.LockOSThread()
	gil := PyGILState_Ensure()
	defer PyGILState_Release(gil)
	defer runtime.UnlockOSThread()

	pyCtxInst := NewPyTreeCallCtxObject(ctx)

	ctx.On("Kill", func() {
		runtime.LockOSThread()
		gil := PyGILState_Ensure()
		defer PyGILState_Release(gil)
		defer runtime.UnlockOSThread()
		kill := pyCtxInst.GetAttrString("kill")
		kill.CallObject(nil)
	})

	argv := PyTuple_New(2)
	PyTuple_SetItem(argv, 0, PyUnicode_FromString(methodName))
	PyTuple_SetItem(argv, 1, pyCtxInst)

	//C.Py_IncRef(toc(pyCtxInst))
	//defer C.Py_DecRef(toc(argv))
	self.branch.CallObject(argv)

	if err := GetPythonError(); err != nil {
		fmt.Println(err)
		ctx.Reject(500, err)
	}
}

//export goObjshTreeAddBranch
func goObjshTreeAddBranch(cbranchObj *C.PyObject,ctreeName *C.PyObject) {
	treeName := PyUnicode_AsUTF8(togo(ctreeName))
	branchObj := togo(cbranchObj)
	if tree, ok := treeExposedToPython[treeName]; ok {
		bname := PyUnicode_AsUTF8(branchObj.GetAttrString("name"))
		pybranch := PyBranch{
			branch: branchObj,
		}
		tree.AddBranchWithName(&pybranch,bname)
	} else {
		fmt.Errorf("Failed to add branch to tree \"%s\"\n", treeName)
	}
}

//export goObjshTreeSureReady
func goObjshTreeSureReady(ctreeName *C.PyObject, cbranchName *C.PyObject) {
	treeName := PyUnicode_AsUTF8(togo(ctreeName))
	branchName := PyUnicode_AsUTF8(togo(cbranchName))
	if tree, ok := treeExposedToPython[treeName]; ok {
		if branch, ok := tree.Branches[branchName]; ok {
			tree.SureReady(branch)
		}
	} else {
		log.Printf("Failed to add branch to tree \"%s\"\n", treeName)
	}
}
