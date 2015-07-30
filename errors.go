package main

import (
	"fmt"
	"log"
	"runtime/debug"
)

type DefaultErrorId string

const (
	E_UNAUTHORIZED           string = "Unauthorized"
	E_INVALID_AUTHENTICATION        = "invalid_authentication"
	E_MISSING_AUTHENTICATION        = "missing_authentication"
	E_INVALID_PRIVILEGES            = "invalid_privileges"
	E_NO_PERMISSION                 = "invalid_permission"
	E_PARSER_ERROR                  = "parse_error"
	E_INVALID_UUID                  = "invalid_uuid"
	E_INVALID_CONTEXT               = "invalid_context"
	E_MISSING_VALUE                 = "salesorder_not_specified"
	E_UNRECOGNIZED_VALUE            = "unrecognized_value"
	E_FILE_NOT_EXISTS               = "file_not_exists"
	E_UNKOWN_ERROR                  = "unknown_error"
)

var (
	deferror *DefaultErrors = NewDefaultErrors()
)

type DefaultErrors struct {
	errorMap map[string]string
}

func NewDefaultErrors() *DefaultErrors {
	r := &DefaultErrors{errorMap: make(map[string]string)}
	r.errorMap[E_INVALID_AUTHENTICATION] = "Invalid authentication"
	r.errorMap[E_MISSING_AUTHENTICATION] = "Missing authentication"
	r.errorMap[E_INVALID_PRIVILEGES] = "Invalid privileges for current user"
	r.errorMap[E_NO_PERMISSION] = "You do not have permission to perform this operation"
	r.errorMap[E_PARSER_ERROR] = "Parser returned an error"
	r.errorMap[E_INVALID_UUID] = "UUID supplied was invalid"
	r.errorMap[E_INVALID_CONTEXT] = "Invalid user context"
	r.errorMap[E_MISSING_VALUE] = "Missing value: %s"
	r.errorMap[E_UNRECOGNIZED_VALUE] = "Unrecognized value %s for %s"
	r.errorMap[E_FILE_NOT_EXISTS] = "File `%s` does not exist"
	r.errorMap[E_UNKOWN_ERROR] = "Operation returned an unknown error"
	return r
}

func (e *DefaultErrors) Get(id ...interface{}) string {
	if len(id) == 0 {
		log.Println("ERROR: Get method was called without a valid id parameter (See stack trace in stderr for details).")
		debug.PrintStack()
		return ""
	}
	key, ok := id[0].(string)
	if ok == false {
		log.Println("ERROR: First parameter of DefaultErrors Get func must be a string value")
		debug.PrintStack()
		return ""
	}
	var data []interface{}
	if len(id) > 1 {
		data = id[1:]
	}
	if m, ok := e.errorMap[key]; ok {
		return fmt.Sprintf(m, data...)
	}
	return key
}
