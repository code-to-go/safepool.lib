package main

/*
typedef struct Result{
	char* err;
    char* data;
} Result;
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"weshare/engine"
	"weshare/model"
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

	var access model.Access
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
