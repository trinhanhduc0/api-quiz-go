package routes

import "github.com/gorilla/mux"

// Router struct để quản lý các route
type Router struct {
	*mux.Router
}

// NewRouter khởi tạo một Router mới
func NewRouter() *Router {
	router := mux.NewRouter()
	return &Router{Router: router}
}
