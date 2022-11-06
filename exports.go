package main

/*
typedef struct Result{
	char* err;
    char* data;
} Result;

typedef struct App {
	void (*feed)(char* name, char* data, int eof);
} App;

#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"weshare/engine"
	"weshare/model"
	"weshare/safe"
	"weshare/transport"
)

//export start
func start() C.Result {
	var res C.Result

	err := engine.Start()
	if err != nil {
		res.err = C.CString(err.Error())
	}
	return res
}

//export stop
func stop() C.Result {
	var res C.Result

	err := engine.Stop()
	if err != nil {
		res.err = C.CString(err.Error())
	}
	return res
}

//export setDomain
func setDomain(domainDef *C.char) *C.char {
	data := C.GoString(domainDef)

	var access model.Transport
	err := json.Unmarshal([]byte(data), &access)
	if err != nil {
		return C.CString(err.Error())
	}

	err = engine.Join(access)
	if err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export saveSafe
func saveSafe(nameC *C.char, configsC *C.char) *C.char {
	name := C.GoString(nameC)
	configsS := C.GoString(configsC)

	var configs []transport.Config
	err := json.Unmarshal([]byte(configsS), &configs)
	if err != nil {
		return C.CString(err.Error())
	}

	if err = safe.Save(name, configs); err != nil {
		return C.CString(err.Error())
	}
	return nil
}

//export openSafe
func openSafe(nameC *C.char, handle *C.int) *C.char {
	name := C.GoString(nameC)
	safe.Load(name)

	return nil
}

//export createSafe
func createSafe(nameDef *C.char, jsonConfig *C.char, handle *C.int) *C.char {
	return nil
}
