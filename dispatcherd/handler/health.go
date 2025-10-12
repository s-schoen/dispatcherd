package handler

import "net/http"

func HandleHealth(w http.ResponseWriter, r *http.Request) error {
	return respondOne(w, r, "OK")
}
