package routes

import (
	"encoding/json"
	"fmt"
	handlers "my_app/internal/handler"
	models "my_app/internal/model"
	"my_app/internal/pkg"
	utils "my_app/internal/util"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type RoutesClass struct {
	r *Router
}

func GetRoutesClass(r *Router) *RoutesClass {
	return &RoutesClass{
		r: r,
	}
}

// InitializeRoutes khởi tạo tất cả các route
func (rc *RoutesClass) InitializeRoutesClass() {
	rc.r.Handle("/class", handlers.AuthMiddleware(http.HandlerFunc(rc.getAllClass))).Methods("GET")
	rc.r.Handle("/class", handlers.AuthMiddleware(http.HandlerFunc(rc.createClass))).Methods("POST")
	rc.r.Handle("/class", handlers.AuthMiddleware(http.HandlerFunc(rc.updateClass))).Methods("PATCH")
	rc.r.Handle("/class", handlers.AuthMiddleware(http.HandlerFunc(rc.deleteClass))).Methods("DELETE")
}

func (r *RoutesClass) createClass(w http.ResponseWriter, req *http.Request) {
	emailID, ok := req.Context().Value("email_id").(string)
	email, ok := req.Context().Value("email").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	var classCreate models.Class

	fieldCreate, err := utils.GenerateUpdateFields(req, &classCreate)
	if err != nil {

		pkg.SendError(w, "Failed to generate update fields", http.StatusInternalServerError)
		return
	}
	if fieldCreate == nil {
		fieldCreate = make(map[string]interface{})
	}

	// Set the email_id in the classCreate struct
	fieldCreate["created_at"] = time.Now()
	fieldCreate["updated_at"] = time.Now()
	fieldCreate["author_mail"] = email
	fieldCreate["email_id"] = emailID

	// Convert question_ids to ObjectID if present
	utils.ConvertIDs(fieldCreate, "test_id")

	// Call the CreateTest function from your models package
	questionRepo := models.NewQuestionRepository("classes")
	create, err := questionRepo.Create(fieldCreate)
	if err != nil {
		pkg.SendError(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	response := bson.M{"_id": create, "class": fieldCreate}
	pkg.SendResponse(w, http.StatusCreated, response)

}

// GetAllClass retrieves all test documents for a specific user based on email_id.
func (r *RoutesClass) getAllClass(w http.ResponseWriter, req *http.Request) {
	// Retrieve the email_id from the request context
	emailID, ok := req.Context().Value("email_id").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("class:%s", emailID)

	if object, err := models.GetRedis().GetObject(cacheKey); err == nil && object != nil {
		pkg.SendResponse(w, http.StatusOK, object)
		return
	}

	// Define the filter for fetching all class associated with the email_id
	filter := bson.M{"email_id": emailID}

	// Retrieve all class for the user
	questionRepo := models.NewQuestionRepository("classes")
	list, err := questionRepo.GetAll(filter)
	if err != nil {
		pkg.SendError(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	if err := models.GetRedis().SetObject(cacheKey, list, 2*time.Minute); err != nil {
		fmt.Println("Failed to cache data in Redis:", err)
	}

	// Return the list of class as JSON
	pkg.SendResponse(w, http.StatusOK, list)
}

func (r *RoutesClass) updateClass(w http.ResponseWriter, req *http.Request) {
	emailID := req.Context().Value("email_id").(string)

	var classUpdate models.Class

	// Generate the fields from the request
	updateFields, err := utils.GenerateUpdate(req, &classUpdate)
	if err != nil {
		pkg.SendError(w, "Failed to generate update fields", http.StatusInternalServerError)
		return
	}

	// Convert question_ids to ObjectID if present
	if test_ids, ok := updateFields["test_id"].([]any); ok && test_ids != nil {
		arrayObject, err := utils.ArrayStringToObjectId(test_ids)
		if err != nil {
			pkg.SendError(w, "Invalid question IDs", http.StatusBadRequest)
			return
		}
		updateFields["test_id"] = arrayObject
	}

	classID, err := utils.StringToObjectId(updateFields["_id"])
	if err != nil {
		pkg.SendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	delete(updateFields, "_id")

	updateFields["updated_at"] = time.Now()

	filter := bson.M{"email_id": emailID, "_id": classID}

	questionRepo := models.NewQuestionRepository("classes")
	update, err := questionRepo.Update(filter, updateFields)
	if err != nil || update.MatchedCount == 0 {
		fmt.Println(err)
		pkg.SendError(w, "Failed to update class", http.StatusInternalServerError)
		return
	}

	response := bson.M{"_id": classID, "test": updateFields}
	// Trả về kết quả cập nhật
	pkg.SendResponse(w, http.StatusOK, response)

}

func (r *RoutesClass) deleteClass(w http.ResponseWriter, req *http.Request) {
	email_id := req.Context().Value("email_id").(string)

	var classCreate struct {
		ID any `json:"_id"`
	}

	if err := json.NewDecoder(req.Body).Decode(&classCreate); err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	test_id, err := utils.StringToObjectId(classCreate.ID)
	if err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Invalid test id", http.StatusBadRequest)
		return
	}
	// Only perform the update if there are fields to update
	questionRepo := models.NewQuestionRepository("classes")
	delete, err := questionRepo.Delete(bson.M{"email_id": email_id, "_id": test_id})
	if err != nil || delete.DeletedCount == 0 {
		pkg.SendError(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	// Return the updated test details
	pkg.SendResponse(w, http.StatusOK, test_id)

}
