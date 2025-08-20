package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"level0/internal/model"
)

func (c *Controller) GetOrder(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	prefix := "/order/"

	if !strings.HasPrefix(path, prefix) {
		log.Printf("Запрос на неккоректную страницу: %s", path)
		http.NotFound(w, r)
		return
	}

	orderUid := strings.TrimPrefix(path, prefix)
	if orderUid == "" {
		http.Error(w, "Не указан айди заказа", http.StatusBadRequest)
		return
	}
	order, err := c.Cache.Get(r.Context(), orderUid)
	if err != nil {
		log.Printf("Ошибка в получении заказа: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "Заказ не найден", http.StatusBadRequest)
		return
	}
	o := model.OrderToResponse(order)
	err = json.NewEncoder(w).Encode(o)
	if err != nil {
		log.Printf("Ошибка кодирования заказа %s: %v", orderUid, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
