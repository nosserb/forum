package handlers

import (
	"log"
	"net/http"
)

type Error struct {
	Message string
	Code    int
	Error   string
}

// Display error with status code, general error type and custom error message
// Write error code to header, log request, and execute error page template
func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	err := templates.ExecuteTemplate(w, "error.html", Error{
		Message: message,
		Code:    statusCode,
		Error:   http.StatusText(statusCode),
	})

	if err != nil {
		log.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
