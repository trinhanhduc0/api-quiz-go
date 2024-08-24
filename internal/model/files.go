package models

import (
	//"context"
	"context"
	"fmt"
	"io"
	"log"
	database "my_app/internal/database"

	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

//func uploadFileHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) {

func UploadFileHandler(w http.ResponseWriter, r *http.Request, email string) string {
	client := database.GetMongoClient()
	r.ParseMultipartForm(10 << 20) // Limit 10MB
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return ""
	}
	defer file.Close()
	createErr := os.MkdirAll("./uploads", os.ModePerm)
	if createErr != nil {
		log.Fatal(createErr)
		return ""
	}

	filePath := "./uploads/" + email + "_" + handler.Filename
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Unable to create the file for writing. Check your write access privilege")
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Println("Error saving the file")
		return ""
	}
	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
	collection := client.Database("dbapp").Collection("files")
	fileData := bson.M{
		"filename":    handler.Filename,
		"file_type":   handler.Header.Get("Content-Type"),
		"file_size":   handler.Size,
		"upload_date": time.Now(),
		"metadata": bson.M{
			"email": email,
		},
	}
	_, err = collection.InsertOne(context.TODO(), fileData)
	if err != nil {
		fmt.Println("Error inserting document:", err)
	}
	return handler.Filename
}

func GetFileHandler(w http.ResponseWriter, r *http.Request, email string, filename string) {
	client := database.GetMongoClient()
	collection := client.Database("dbapp").Collection("files")

	// Tìm kiếm file dựa trên email và filename
	filter := bson.M{
		"filename":       filename,
		"metadata.email": email,
	}
	var fileData bson.M
	err := collection.FindOne(context.TODO(), filter).Decode(&fileData)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Lấy đường dẫn của file đã upload
	filePath := "./uploads/" + email + "_" + filename
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set Content-Type và các headers khác
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", fileData["file_type"].(string))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileData["file_size"].(int64)))

	// Gửi nội dung file tới client
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error writing file to response", http.StatusInternalServerError)
		return
	}
}
