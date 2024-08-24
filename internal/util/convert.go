package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	models "my_app/internal/model"
	"net/http"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArrayStringToObjectID converts a BSON array of string IDs to a slice of ObjectIDs.
func ArrayStringToObjectId(arr []any) ([]primitive.ObjectID, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range arr {
		strID, ok := id.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", id)
		}
		objectID, err := primitive.ObjectIDFromHex(strID)
		if err != nil {
			return nil, fmt.Errorf("invalid ObjectID format: %v", err)
		}
		objectIDs = append(objectIDs, objectID)
	}
	return objectIDs, nil
}

func StringToObjectId(id any) (any, error) {
	objectID, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		return primitive.NilObjectID, err // Return an error instead of panicking
	}
	return objectID, nil
}

func StringToTime(timeStr string) (time.Time, error) {
	// Chuyển đổi chuỗi thời gian theo định dạng ISO 8601
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %v", err)
	}
	return t, nil
}

// GenerateUpdateFields generates the fields to update for MongoDB based on the non-empty fields in the provided struct
func GenerateUpdateFields(req *http.Request, targetStruct any) (bson.M, error) {
	// Decode the JSON body into the provided struct
	if err := json.NewDecoder(req.Body).Decode(targetStruct); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(targetStruct)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("targetStruct must be a pointer")
	}

	v = v.Elem() // Dereference the pointer to get the underlying struct
	typeOfStruct := v.Type()
	updateFields := bson.M{}

	// Iterate over the fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typeOfStruct.Field(i).Tag.Get("json")
		if fieldName == "" {
			fieldName = typeOfStruct.Field(i).Name
		}

		// Check if the field has a non-empty value
		switch field.Kind() {
		case reflect.String:
			if field.String() != "" {
				updateFields[fieldName] = field.String()
			}
		case reflect.Slice:
			updateFields[fieldName] = field.Interface()
		case reflect.Ptr:
			if !field.IsNil() {
				updateFields[fieldName] = field.Elem().Interface()
			}
		case reflect.Struct:
			// Handle nested structs like question_content
			nestedFields, err := generateNestedFields(field)
			if err != nil {
				return nil, err
			}
			if len(nestedFields) > 0 {
				updateFields[fieldName] = nestedFields
			}
		case reflect.Int:
			updateFields[fieldName] = field.Int()
		case reflect.Bool:
			updateFields[fieldName] = field.Bool()
		}
	}

	// Return error if no fields are present
	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	return updateFields, nil
}

// GenerateUpdateFields generates the fields to update for MongoDB based on the non-empty fields in the provided struct
func GenerateUpdate(req *http.Request, targetStruct any) (bson.M, error) {
	// Decode the JSON body into the provided struct
	if err := json.NewDecoder(req.Body).Decode(targetStruct); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(targetStruct)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("targetStruct must be a pointer")
	}

	v = v.Elem() // Dereference the pointer to get the underlying struct
	typeOfStruct := v.Type()
	updateFields := bson.M{}

	// Iterate over the fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typeOfStruct.Field(i).Tag.Get("json")
		if fieldName == "" {
			fieldName = typeOfStruct.Field(i).Name
		}

		// Check if the field has a non-empty value
		switch field.Kind() {
		case reflect.String:
			if field.String() != "" {
				updateFields[fieldName] = field.String()
			}
		case reflect.Slice:
			if field.Len() > 0 {
				updateFields[fieldName] = field.Interface()
			}
		case reflect.Ptr:
			if !field.IsNil() {
				updateFields[fieldName] = field.Elem().Interface()
			}
		case reflect.Struct:
			// Handle nested structs like question_content
			nestedFields, err := generateNestedFields(field)
			if err != nil {
				return nil, err
			}
			if len(nestedFields) > 0 {
				updateFields[fieldName] = nestedFields
			}
		case reflect.Int:
			updateFields[fieldName] = field.Int()
		case reflect.Bool:
			fmt.Println(field)
			updateFields[fieldName] = field.Bool()
		}
	}

	// Return error if no fields are present
	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	return updateFields, nil
}

// Generate fields for nested structs
func generateNestedFields(field reflect.Value) (bson.M, error) {
	nestedFields := bson.M{}
	for j := 0; j < field.NumField(); j++ {
		nestedField := field.Field(j)
		nestedFieldType := field.Type().Field(j) //
		nestedFieldName := field.Type().Field(j).Tag.Get("json")
		if nestedFieldName == "" {
			nestedFieldName = field.Type().Field(j).Name
		}
		// Skip unexported fields
		if !nestedFieldType.IsExported() { //
			continue
		}
		if !isEmpty(nestedField) {
			nestedFields[nestedFieldName] = nestedField.Interface()
		}
	}
	return nestedFields, nil
}

// Check if a value is empty
func isEmpty(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Slice:
		return value.Len() == 0
	case reflect.Ptr:
		return value.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
	case reflect.Int:
		return value.Interface() == reflect.Zero(value.Type()).Interface()
	case reflect.Bool:
		return value.Bool()
	}

	return false
}
func BuildBSON(testAnswer models.TestAnswer) bson.M {
	bsonData := bson.M{
		"test_id":    testAnswer.TestId,
		"email_id":   testAnswer.EmailID,
		"email":      testAnswer.Email,
		"created_at": testAnswer.CreatedAt,
	}

	var questions []bson.M
	for _, qa := range testAnswer.ListQuestionAnswer {
		qaData := bson.M{
			"question_id": qa.QuestionID,
		}

		if len(qa.FillInTheBlanks) > 0 {
			qaData["fill_in_the_blank"] = qa.FillInTheBlanks
		}

		if len(qa.Options) > 0 {
			var options []bson.M
			for _, opt := range qa.Options {
				optData := bson.M{
					"text":      opt.Text,
					"image_url": opt.ImageURL,
				}

				// Add each field conditionally
				if opt.Match != "" {
					optData["match"] = opt.Match
				} else if opt.Order != 0 {
					optData["order"] = opt.Order
				} else {
					optData["iscorrect"] = opt.IsCorrect
				}

				options = append(options, optData)
			}
			qaData["options"] = options
		}

		questions = append(questions, qaData)
	}
	bsonData["question_answer"] = questions

	return bsonData
}

// ConvertIDs converts string IDs to ObjectIDs for specified fields in a map.
func ConvertIDs(fields map[string]interface{}, fieldNames ...string) {
	for _, fieldName := range fieldNames {
		if ids, ok := fields[fieldName].([]any); ok && ids != nil {
			if arrayObject, err := ArrayStringToObjectId(ids); err == nil {
				fields[fieldName] = arrayObject
			} else {
				fmt.Printf("Error converting IDs: %v\n", err)
			}
		}
	}
}

// RemoveKeysFromList removes specified keys from each map in the list
func RemoveKeysFromList(list []bson.M, keysToRemove []string) {
	for i := range list {
		for _, key := range keysToRemove {
			delete(list[i], key)
		}
	}
}

// RemoveAnswer removes specific fields from elements in an array of bson.M
func RemoveAnswer(questionList []bson.M, answerField string) {
	for _, question := range questionList {
		delete(question, answerField)
	}
}

// convertToBsonMArray chuyển đổi bson.A sang []bson.M
func ConvertToBsonMArray(array bson.A) []bson.M {
	var convertedArray []bson.M
	for _, item := range array {
		if doc, ok := item.(bson.M); ok {
			convertedArray = append(convertedArray, doc)
		}
	}
	return convertedArray
}

func RemoveEmptyFillInTheBlanks(fillInTheBlanks []models.FillInTheBlank) []models.FillInTheBlank {
	var result []models.FillInTheBlank
	for _, item := range fillInTheBlanks {
		if item.CorrectAnswer != "" { // Example condition, adjust as needed
			result = append(result, item)
		}
	}
	return result
}

func RemoveEmptyOptions(options []models.Option) []models.Option {
	var result []models.Option
	for _, option := range options {
		if option.Text != "" || option.ImageURL != "" || option.Match != "" { // Adjust conditions as needed
			result = append(result, option)
		}
	}
	return result
}

// Function to remove empty QuestionAnswer entries
func RemoveEmptyQuestionAnswers(answers []models.QuestionAnswer) []models.QuestionAnswer {
	var result []models.QuestionAnswer
	for _, answer := range answers {
		// Clean up the fields
		answer.FillInTheBlanks = RemoveEmptyFillInTheBlanks(answer.FillInTheBlanks)
		answer.Options = RemoveEmptyOptions(answer.Options)

		// Only include non-empty QuestionAnswer entries
		if answer.QuestionID != "" && len(answer.FillInTheBlanks) > 0 || len(answer.Options) > 0 {
			result = append(result, answer)
		}
	}
	return result
}
