package api

import (
	"fmt"
	"net/http"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	res := "Hot reload is working!"
	fmt.Fprintln(w, res)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement test logic
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"message": "Test endpoint - implementation pending"}`)
}
