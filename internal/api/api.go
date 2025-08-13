package api

import (
	"Level0/internal/api/controllers"
	"net/http"
)

func RouteController(c *controllers.Controller) http.Handler {
	mux := http.NewServeMux()
	// Для простоты сервер будет раздавать и статику
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/info/", http.StripPrefix("/info/", fs))
	mux.HandleFunc("/order/", c.GetOrder)
	return mux
}
