package initialize

import (
	"log"
	"my_app/internal/routes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Router struct để quản lý các route
type Router struct {
	*mux.Router
}

func InitRouter() {
	// Initialize CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	router := routes.NewRouter()
	routes.GetRoutesAuth(router).InitializeRoutesAuth()
	routes.GetRoutesUser(router).InitializeRoutesUser()
	routes.GetRoutesClass(router).InitializeRoutesClass()
	routes.GetRoutesQuestion(router).InitializeRoutesQuestion()
	routes.GetRoutesAuthorTest(router).InitializeRoutesManageTests()
	routes.GetRoutesFile(router).InitializeRoutesFile()
	routes.GetRoutesTest(router).InitializeRoutesTests()

	// Apply CORS handler
	handler := c.Handler(router)

	// Start the server
	port := "8080"
	log.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
