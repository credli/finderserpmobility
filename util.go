package main

import (
	"net/http"
	"os"
)

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
