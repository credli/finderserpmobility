package main

import (
	"github.com/pborman/uuid"
	"net/http"
	"os"
)

func toLittleEndian(largeEndian uuid.UUID) uuid.UUID {
	littleEndian := uuid.NewUUID()
	for i := 8; i < 16; i++ {
		littleEndian[i] = largeEndian[i]
	}
	littleEndian[3] = largeEndian[0]
	littleEndian[2] = largeEndian[1]
	littleEndian[1] = largeEndian[2]
	littleEndian[0] = largeEndian[3]
	littleEndian[5] = largeEndian[4]
	littleEndian[4] = largeEndian[5]
	littleEndian[7] = largeEndian[6]
	littleEndian[6] = largeEndian[7]
	return littleEndian
}

func toLargeEndian(littleEndian uuid.UUID) uuid.UUID {
	largeEndian := uuid.NewUUID()
	for i := 8; i < 16; i++ {
		largeEndian[i] = littleEndian[i]
	}
	largeEndian[0] = littleEndian[3]
	largeEndian[1] = littleEndian[2]
	largeEndian[2] = littleEndian[1]
	largeEndian[3] = littleEndian[0]
	largeEndian[4] = littleEndian[5]
	largeEndian[5] = littleEndian[4]
	largeEndian[6] = littleEndian[7]
	largeEndian[7] = littleEndian[6]
	return largeEndian
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, model interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
