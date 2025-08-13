package controllers

import (
	"fmt"
	"net/http"
)

func (c *Controller) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello!")
}
