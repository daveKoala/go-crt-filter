package api

import (
	"fmt"
	"net/http"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	res := "Hot reload is working!"
	fmt.Fprintln(w, res)
}
