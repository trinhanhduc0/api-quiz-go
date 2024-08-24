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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoutesQuestion struct {
	r *Router
}

func GetRoutesQuestion(r *Router) *RoutesQuestion {
	return &RoutesQuestion{
		r: r,
	}
}

func (rq *RoutesQuestion) InitializeRoutesQuestion() {
	rq.r.Handle("/questions", handlers.AuthMiddleware(http.HandlerFunc(rq.createQuestions))).Methods("POST")
	rq.r.Handle("/questions", handlers.AuthMiddleware(http.HandlerFunc(rq.getAllQuestions))).Methods("GET")
	rq.r.Handle("/questions", handlers.AuthMiddleware(http.HandlerFunc(rq.updateQuestion))).Methods("PATCH")
	rq.r.Handle("/questions", handlers.AuthMiddleware(http.HandlerFunc(rq.deleteQuestion))).Methods("DELETE")
}

func (r *RoutesQuestion) createQuestions(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("email_id").(string)

	var question models.Question

	// Generate the update fields from the request
	questionField, err := utils.GenerateUpdateFields(req, &question)
	if err != nil {
		fmt.Println(err)
		pkg.SendError(w, "Invalid field create", http.StatusBadRequest)
		return
	}

	fmt.Printf("Type of questionField['options']: %T\n", questionField["options"])

	if options, ok := questionField["options"].([]models.Option); ok {
		for i := range options {
			// Assign a new ObjectID to each option
			options[i].ID = primitive.NewObjectID()
		}
		questionField["options"] = options // Update the options field with new _id values
	}

	// Add metadata and timestamps
	questionField["metadata"] = bson.M{"author": userID}
	questionField["create_at"] = time.Now()
	questionField["updated_at"] = time.Now()

	// Insert the new question into the database
	questionRepo := models.NewQuestionRepository("questions")
	insertedID, err := questionRepo.Create(questionField)
	if err != nil {
		pkg.SendError(w, "Question not created", http.StatusInternalServerError)
		return
	}

	// Send a successful response with the inserted ID
	response := bson.M{"_id": insertedID, "question": questionField}
	pkg.SendResponse(w, http.StatusCreated, response)
}

func (r *RoutesQuestion) getAllQuestions(w http.ResponseWriter, req *http.Request) {
	userID := req.Context().Value("email_id").(string)

	filter := bson.M{"metadata.author": userID}

	questionRepo := models.NewQuestionRepository("questions")
	questions, err := questionRepo.GetAll(filter)
	if err != nil {
		pkg.SendError(w, "Failed to get questions", http.StatusInternalServerError)
		return
	}
	pkg.SendResponse(w, http.StatusOK, questions)
}

func (r *RoutesQuestion) updateQuestion(w http.ResponseWriter, req *http.Request) {
	emailID := req.Context().Value("email_id").(string)

	var question models.Question

	questionField, err := utils.GenerateUpdateFields(req, &question)

	if err != nil {
		pkg.SendError(w, "Invalid field", http.StatusBadRequest)
		return
	}

	objectID, err := utils.StringToObjectId(questionField["_id"].(string))
	if err != nil {
		pkg.SendError(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	// questionField["metadata"] = metadata
	questionField["updated_at"] = time.Now()

	if options, ok := questionField["options"].([]models.Option); ok {
		for i := range options {
			// Assign a new ObjectID to each option
			options[i].ID = primitive.NewObjectID()
		}
		questionField["options"] = options // Update the options field with new _id values
	}

	filter := bson.M{"_id": objectID, "metadata.author": emailID}

	delete(questionField, "_id")
	questionRepo := models.NewQuestionRepository("questions")
	result, err := questionRepo.Update(filter, questionField)

	if err != nil || result.MatchedCount == 0 {
		pkg.SendError(w, "Failed to update question", http.StatusInternalServerError)
		return
	}

	response := bson.M{"_id": objectID, "question": question}

	pkg.SendResponse(w, http.StatusOK, response)

}

func (r *RoutesQuestion) deleteQuestion(w http.ResponseWriter, req *http.Request) {
	emailID := req.Context().Value("email_id").(string)

	var question bson.M
	if err := json.NewDecoder(req.Body).Decode(&question); err != nil {
		pkg.SendError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_id, ok := question["_id"].(string)
	if !ok {
		pkg.SendError(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		pkg.SendError(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID, "metadata.author": emailID}

	questionRepo := models.NewQuestionRepository("questions")
	result, err := questionRepo.Delete(filter)
	if err != nil {
		pkg.SendError(w, "Failed to delete question", http.StatusInternalServerError)
		return
	}

	pkg.SendResponse(w, http.StatusInternalServerError, result)
}
