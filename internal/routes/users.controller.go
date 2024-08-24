package routes

import (
	"encoding/json"
	"fmt"
	handlers "my_app/internal/handler"
	models "my_app/internal/model"
	"my_app/internal/pkg"
	utils "my_app/internal/util"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

type RoutesUser struct {
	r *Router
}

func GetRoutesUser(r *Router) *RoutesUser {
	return &RoutesUser{
		r: r,
	}
}

// InitializeRoutes khởi tạo tất cả các route
func (rs *RoutesUser) InitializeRoutesUser() {
	rs.r.Handle("/users", handlers.AuthMiddleware(http.HandlerFunc(rs.getUsers))).Methods("GET")
	rs.r.Handle("/users", handlers.AuthMiddleware(http.HandlerFunc(rs.checkToken))).Methods("POST")
	rs.r.Handle("/users", handlers.AuthMiddleware(http.HandlerFunc(rs.updateUser))).Methods("PATCH")
	rs.r.Handle("/users", handlers.AuthMiddleware(http.HandlerFunc(rs.deleteUser))).Methods("DELETE")
}

func (r *RoutesUser) getUsers(w http.ResponseWriter, req *http.Request) {
	// Lấy giá trị từ context
	email_id, ok := req.Context().Value("email_id").(string)
	if !ok {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	repoUser := models.NewQuestionRepository("users")
	user_row, err := repoUser.GetFilter(bson.M{"email_id": email_id})
	// Nếu không tìm thấy người dùng
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Trả về thông tin người dùng
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user_row); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *RoutesUser) checkToken(w http.ResponseWriter, req *http.Request) {
	// Lấy giá trị từ context
	email, ok := req.Context().Value("email").(string)
	if !ok {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}
	pkg.SendResponse(w, http.StatusOK, bson.M{"email": email})
}

func (r *RoutesUser) updateUser(w http.ResponseWriter, req *http.Request) {
	emailID, ok := req.Context().Value("email_id").(string)
	if !ok {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	var userUpdate models.User

	// Generate update fields from the request body
	updateFields, err := utils.GenerateUpdateFields(req, &userUpdate)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// Check if there are any fields to update
	if len(updateFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	hashPassword := utils.HashPassword(updateFields["password"].(string))

	updateFields["password"] = hashPassword

	// Create a repository for users and perform the update
	userRepo := models.NewQuestionRepository("users")
	update, err := userRepo.Update(bson.M{"email_id": emailID}, updateFields)

	if err != nil || update.MatchedCount == 0 {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	pkg.SendResponse(w, http.StatusOK, update)
}

func (r *RoutesUser) deleteUser(w http.ResponseWriter, req *http.Request) {
	// Code để xóa người dùng theo ID
	vars := mux.Vars(req)
	id := vars["id"]
	w.Write([]byte("Delete user by ID: " + id))
}
