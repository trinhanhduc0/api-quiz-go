package routes

import (
	"context"
	"log"
	handlers "my_app/internal/handler"
	models "my_app/internal/model"
	"net/http"
)

type RoutesAuth struct {
	r *Router
}

func GetRoutesAuth(r *Router) *RoutesAuth {
	return &RoutesAuth{
		r: r,
	}
}

// InitializeRoutes khởi tạo tất cả các route
func (ra *RoutesAuth) InitializeRoutesAuth() {
	ra.r.Handle("/api/google/login", http.HandlerFunc(ra.login)).Methods("POST")
}

func (r *RoutesAuth) login(w http.ResponseWriter, req *http.Request) {
	authClient, err := models.GetAuth()
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth client: %v", err)
		return
	}
	// Add custom handler for Google login to the router
	handlers.Login(w, req, authClient, context.TODO())
}
