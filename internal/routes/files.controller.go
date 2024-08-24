package routes

import (
	"encoding/json"
	handlers "my_app/internal/handler"
	models "my_app/internal/model"

	"net/http"
)

type RoutesFile struct {
	r *Router
}

func GetRoutesFile(r *Router) *RoutesFile {
	return &RoutesFile{
		r: r,
	}
}

// InitializeRoutes khởi tạo tất cả các route
func (rf *RoutesFile) InitializeRoutesFile() {
	rf.r.Handle("/file", handlers.AuthMiddleware(http.HandlerFunc(rf.getFile))).Methods("POST")
	rf.r.Handle("/upfile", handlers.AuthMiddleware(http.HandlerFunc(rf.uploadFile))).Methods("POST")
}

func (r *RoutesFile) uploadFile(w http.ResponseWriter, req *http.Request) {
	email, ok := req.Context().Value("email").(string)
	if !ok {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}
	models.UploadFileHandler(w, req, email)
}

func (r *RoutesFile) getFile(w http.ResponseWriter, req *http.Request) {
	// email_id, ok := req.Context().Value("email_id").(string)
	// if !ok {
	// 	pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
	// 	return
	// }
	var file struct {
		FileName string `json:"file_url"`
		Email    string `json:"email"`
	}

	// Decode request body
	if err := json.NewDecoder(req.Body).Decode(&file); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	models.GetFileHandler(w, req, file.Email, file.FileName)
}
