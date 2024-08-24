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

type RoutesAuthorTest struct {
	r *Router
}

func GetRoutesAuthorTest(r *Router) *RoutesAuthorTest {
	return &RoutesAuthorTest{
		r: r,
	}
}

// InitializeRoutesTests initializes all test-related routes
func (rat *RoutesAuthorTest) InitializeRoutesManageTests() {
	rat.r.Handle("/tests", handlers.AuthMiddleware(http.HandlerFunc(rat.getAllTestFromAuthor))).Methods("GET")
	rat.r.Handle("/tests", handlers.AuthMiddleware(http.HandlerFunc(rat.createTest))).Methods("POST")
	rat.r.Handle("/tests", handlers.AuthMiddleware(http.HandlerFunc(rat.updateTest))).Methods("PATCH")
	rat.r.Handle("/tests", handlers.AuthMiddleware(http.HandlerFunc(rat.deleteTest))).Methods("DELETE")
}

func (r *RoutesAuthorTest) createTest(w http.ResponseWriter, req *http.Request) {
	emailID := req.Context().Value("email_id").(string)
	var test models.Test

	// Generate update fields from the test struct
	testField, err := utils.GenerateUpdateFields(req, &test)
	if err != nil {
		pkg.SendError(w, "Failed to generate update fields", http.StatusInternalServerError)
		return
	}

	testField["created_at"] = time.Now()
	testField["updated_at"] = time.Now()
	testField["email_id"] = emailID
	if startTime, ok := testField["start_time"]; ok {
		testField["start_time"], err = utils.StringToTime(startTime.(string))
		if err != nil {
			// Handle the error if needed
		}
	}

	if endTime, ok := testField["end_time"]; ok {
		testField["end_time"], err = utils.StringToTime(endTime.(string))
		if err != nil {
			// Handle the error if needed
		}
	}
	utils.ConvertIDs(testField, "question_ids", "class_ids")

	testRepo := models.NewQuestionRepository("tests")
	created, err := testRepo.Create(testField)
	if err != nil {
		pkg.SendError(w, "Failed to create test", http.StatusInternalServerError)
		return
	}

	pkg.SendResponse(w, http.StatusCreated, bson.M{"_id": created, "data": testField})
}

func (r *RoutesAuthorTest) getAllTestFromAuthor(w http.ResponseWriter, req *http.Request) {
	emailID, ok := req.Context().Value("email_id").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("test:%s", emailID)

	// Attempt to get data from Redis
	if object, err := models.GetRedis().GetObject(cacheKey); err == nil && object != nil {
		pkg.SendResponse(w, http.StatusOK, object)
		return
	}

	filter := bson.M{"email_id": emailID}

	questionRepo := models.NewQuestionRepository("tests")
	list, err := questionRepo.GetAll(filter)
	if err != nil {
		pkg.SendError(w, "Failed to retrieve tests", http.StatusInternalServerError)
		return
	}

	// Cache the result in Redis
	if err := models.GetRedis().SetObject(cacheKey, list, 2*time.Minute); err != nil {
		fmt.Println("Failed to cache data in Redis:", err)
	}

	pkg.SendResponse(w, http.StatusOK, list)
}

func (r *RoutesAuthorTest) updateTest(w http.ResponseWriter, req *http.Request) {
	emailID, ok := req.Context().Value("email_id").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}
	var testUpdate models.Test

	updateFields, err := utils.GenerateUpdateFields(req, &testUpdate)
	if err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Failed to generate update fields", http.StatusInternalServerError)
		return
	}

	utils.ConvertIDs(updateFields, "question_ids", "class_ids")

	testID, err := utils.StringToObjectId(updateFields["_id"])
	if err != nil {
		pkg.SendError(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	delete(updateFields, "_id")

	updateFields["updated_at"] = time.Now()
	if startTime, ok := updateFields["start_time"]; ok {
		updateFields["start_time"], err = utils.StringToTime(startTime.(string))
		if err != nil {
			// Handle the error if needed
		}
	}

	if endTime, ok := updateFields["end_time"]; ok {
		updateFields["end_time"], err = utils.StringToTime(endTime.(string))
		if err != nil {
			// Handle the error if needed
		}
	}

	filter := bson.M{"email_id": emailID, "_id": testID}

	testRepo := models.NewQuestionRepository("tests")
	update, err := testRepo.Update(filter, updateFields)
	if err != nil || update.MatchedCount == 0 {
		pkg.SendError(w, "Failed to update test", http.StatusInternalServerError)
		return
	}

	pkg.SendResponse(w, http.StatusOK, bson.M{"_id": testID, "test": updateFields})
}

func (r *RoutesAuthorTest) deleteTest(w http.ResponseWriter, req *http.Request) {
	emailID := req.Context().Value("email_id").(string)
	var testUpdate struct {
		ID any `json:"_id"`
	}

	if err := json.NewDecoder(req.Body).Decode(&testUpdate); err != nil {
		pkg.SendError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	testID, err := utils.StringToObjectId(testUpdate.ID)
	if err != nil {
		pkg.SendError(w, "Failed to convert test_id", http.StatusInternalServerError)
		return
	}

	questionRepo := models.NewQuestionRepository("tests")
	result, err := questionRepo.Delete(bson.M{"_id": testID, "email_id": emailID})
	if err != nil || result.DeletedCount == 0 {
		pkg.SendError(w, "Failed to delete test", http.StatusInternalServerError)
		return
	}

	pkg.SendResponse(w, http.StatusOK, result)
}
