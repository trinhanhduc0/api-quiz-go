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

type RoutesTest struct {
	r *Router
}

func GetRoutesTest(r *Router) *RoutesTest {
	return &RoutesTest{
		r: r,
	}
}

// InitializeRoutesTests initializes all test-related routes
func (rt *RoutesTest) InitializeRoutesTests() {
	rt.r.Handle("/gettests", handlers.AuthMiddleware(http.HandlerFunc(rt.getAllTestByEmail))).Methods("POST")
	rt.r.Handle("/getclasses", handlers.AuthMiddleware(http.HandlerFunc(rt.getAllClassByEmail))).Methods("GET")
	rt.r.Handle("/getquestions", handlers.AuthMiddleware(http.HandlerFunc(rt.getQuestionOfTest))).Methods("POST")
	rt.r.Handle("/donetest", handlers.AuthMiddleware(http.HandlerFunc(rt.getDoneTest))).Methods("POST")
	rt.r.Handle("/sendtest", handlers.AuthMiddleware(http.HandlerFunc(rt.sendTest))).Methods("POST")
}

// GetAllClass retrieves all test documents for a specific user based on email_id.
func (r *RoutesTest) getAllClassByEmail(w http.ResponseWriter, req *http.Request) {
	// Retrieve the email from the request context
	email, ok := req.Context().Value("email").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	// Define the filter for fetching all class associated with the email_id
	filter := bson.M{"students_accept": email}

	// Retrieve all class for the user
	questionRepo := models.NewQuestionRepository("classes")
	list, err := questionRepo.GetAll(filter)
	if err != nil {
		pkg.SendError(w, "Failed to get class", http.StatusInternalServerError)
		return
	}

	utils.RemoveKeysFromList(list, []string{"email_id", "students_wait", "students_accept"})

	// Return the list of class as JSON
	pkg.SendResponse(w, http.StatusOK, list)
}

func (r *RoutesTest) getAllTestByEmail(w http.ResponseWriter, req *http.Request) {
	email, ok := req.Context().Value("email").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}
	// Decode request body
	var class_id bson.M
	if err := json.NewDecoder(req.Body).Decode(&class_id); err != nil {
		pkg.SendError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	cacheKey := fmt.Sprintf("class:%s", class_id["_id"])
	fmt.Println(cacheKey)
	// Try to get data from Redis cache
	if object, err := models.GetRedis().GetObject(cacheKey); err == nil {
		for _, v := range object["allowed_users"].([]any) {
			if v == email {
				delete(object, "allowed_users")
				pkg.SendResponse(w, http.StatusOK, object["test"])
				return
			}
		}
	}

	var err error
	class_id["_id"], err = utils.StringToObjectId(class_id["_id"])

	filter := bson.M{"class_ids": class_id["_id"], "allowed_users": email}

	testRepo := models.NewQuestionRepository("tests")

	list, err := testRepo.GetAll(filter)
	if err != nil {
		pkg.SendError(w, "Failed to retrieve tests", http.StatusInternalServerError)
		return
	}

	filter = bson.M{"_id": class_id["_id"]}
	classRepo := models.NewQuestionRepository("classes")
	infor_class, err := classRepo.GetFilter(filter)
	if err != nil {
		pkg.SendError(w, "Failed to get test", http.StatusInternalServerError)
		return
	}
	utils.RemoveKeysFromList(list, []string{"allowed_users", "email_id"})

	allTest := bson.M{"allowed_users": infor_class["students_accept"], "test": list}

	// Cache the result
	if err := models.GetRedis().SetObject(cacheKey, allTest, 5*time.Minute); err != nil {
		fmt.Println("Failed to cache data:", err)
	}

	pkg.SendResponse(w, http.StatusOK, list)
}

func (r *RoutesTest) getQuestionOfTest(w http.ResponseWriter, req *http.Request) {
	// Extract email from context
	email, ok := req.Context().Value("email").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var detail bson.M
	if err := json.NewDecoder(req.Body).Decode(&detail); err != nil {
		pkg.SendError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	idTest, err := utils.StringToObjectId(detail["_id"].(string))
	if err != nil {
		pkg.SendError(w, "Failed to retrieve test ID", http.StatusInternalServerError)
		return
	}
	cacheKey := fmt.Sprintf("test:%s", idTest)

	// Try to get data from Redis cache
	if object, err := models.GetRedis().GetObject(cacheKey); err == nil && object != nil {
		for _, v := range object["allowed_users"].([]any) {
			if v == email {
				delete(object, "allowed_users")
				pkg.SendResponse(w, http.StatusOK, object["questions"])
				return
			}
		}
	}

	// Retrieve test details from MongoDB
	filter := bson.M{"_id": idTest, "allowed_users": email}
	projection := bson.M{"question_ids": 1, "is_test": 1, "allowed_users": 1, "_id": 0}

	testRepo := models.NewQuestionRepository("tests")
	testDetails, err := testRepo.GetWithProjection(filter, projection)
	if err != nil {
		pkg.SendError(w, "Failed to retrieve test details", http.StatusInternalServerError)
		return
	}

	// Retrieve questions
	questionFilter := bson.M{"_id": bson.M{"$in": testDetails["question_ids"]}}
	questionRepo := models.NewQuestionRepository("questions")
	questionList, err := questionRepo.GetAll(questionFilter)
	if err != nil {
		pkg.SendError(w, "Failed to retrieve questions", http.StatusInternalServerError)
		return
	}

	// Process questions if it's a test
	if isTest, ok := testDetails["is_test"].(bool); ok && isTest {
		for i := range questionList {
			pkg.ProcessQuestion(questionList[i])
		}
	}

	// Remove unnecessary keys
	utils.RemoveKeysFromList(questionList, []string{"metadata"})
	allTest := bson.M{"allowed_users": testDetails["allowed_users"], "questions": questionList}
	// Cache the result
	if err := models.GetRedis().SetObject(cacheKey, allTest, 5*time.Minute); err != nil {
		fmt.Println("Failed to cache data:", err)
	}

	// Send response
	pkg.SendResponse(w, http.StatusOK, questionList)
}

func (r *RoutesTest) sendTest(w http.ResponseWriter, req *http.Request) {
	//Extract email from context
	email, ok := req.Context().Value("email").(string)
	email_id, ok := req.Context().Value("email_id").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var test_answer models.TestAnswer

	if err := json.NewDecoder(req.Body).Decode(&test_answer); err != nil {
		pkg.SendError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	for i := range test_answer.ListQuestionAnswer {
		questionAnswer := &test_answer.ListQuestionAnswer[i]

		objectID, err := utils.StringToObjectId(questionAnswer.QuestionID)
		if err != nil {
			// Handle the error appropriately; for example, log it and continue
			fmt.Printf("Error converting QuestionID to ObjectID: %v\n", err)
			continue // or handle the error according to your needs
		}
		questionAnswer.QuestionID = objectID

		// Remove empty fields in QuestionAnswercannot use utils.StringToObjectId(test_answer.TestId) (value of type any) as string value in assignment: need
		questionAnswer.FillInTheBlanks = utils.RemoveEmptyFillInTheBlanks(questionAnswer.FillInTheBlanks)
		questionAnswer.Options = utils.RemoveEmptyOptions(questionAnswer.Options)
	}
	test_answer.Email = email
	test_answer.EmailID = email_id
	test_answer.CreatedAt = time.Now()
	var err error
	test_answer.TestId, err = utils.StringToObjectId(test_answer.TestId)
	if err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Invalid test ID", http.StatusInternalServerError)
		return
	}
	// Convert TestAnswer to BSON
	bsonData := utils.BuildBSON(test_answer)

	testRepo := models.NewQuestionRepository("answers")
	created, err := testRepo.Create(bsonData)
	if err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Failed to send test", http.StatusInternalServerError)
		return
	}

	pkg.SendResponse(w, http.StatusOK, created)

}

func (r *RoutesTest) getDoneTest(w http.ResponseWriter, req *http.Request) {
	email_id, ok := req.Context().Value("email_id").(string)
	if !ok {
		pkg.SendError(w, "Invalid email ID", http.StatusBadRequest)
		return
	}
	var test_id bson.M

	if err := json.NewDecoder(req.Body).Decode(&test_id); err != nil {
		pkg.SendError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var err error
	test_id["test_id"], err = utils.StringToObjectId(test_id["test_id"])
	if err != nil {
		pkg.SendError(w, "Invalid Test ID", http.StatusBadRequest)
		return
	}

	fillter := bson.M{"email_id": email_id, "test_id": test_id["test_id"]}

	testRepo := models.NewQuestionRepository("answers")
	list, err := testRepo.GetAll(fillter)
	if err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Failed to send test", http.StatusInternalServerError)
		return
	}
	pkg.SendResponse(w, http.StatusOK, list)

}
