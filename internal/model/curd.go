package models

import (
	"context"
	database "my_app/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CRUDOperations interface {
	Create(any) (primitive.ObjectID, error)
	GetAll(bson.M) ([]bson.M, error)
	GetWithProjection(bson.M, bson.M) []bson.M
	GetFilter(bson.M) ([]bson.M, error)
	Update(bson.M, bson.M) (*mongo.UpdateResult, error)
	Delete(bson.M) (*mongo.DeleteResult, error)
}

type QuestionRepository struct {
	Collection *mongo.Collection
}

func NewQuestionRepository(coll string) *QuestionRepository {
	client := database.GetMongoClient()
	db := client.Database("dbapp")
	return &QuestionRepository{
		Collection: db.Collection(coll),
	}
}

func (r *QuestionRepository) Create(question any) (primitive.ObjectID, error) {
	result, err := r.Collection.InsertOne(context.TODO(), question)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *QuestionRepository) GetAll(filter bson.M) ([]bson.M, error) {
	cursor, err := r.Collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	for cursor.Next(context.TODO()) {
		var question bson.M
		if err := cursor.Decode(&question); err != nil {
			return nil, err
		}
		results = append(results, question)
	}
	return results, nil
}

// GetWithProjection trả về một tài liệu với projection
func (r *QuestionRepository) GetWithProjection(filter bson.M, projection bson.M) (bson.M, error) {
	var result bson.M
	err := r.Collection.FindOne(context.TODO(), filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Không tìm thấy tài liệu
		}
		return nil, err // Lỗi khác
	}
	return result, nil
}

func (r *QuestionRepository) GetFilter(filter bson.M) (bson.M, error) {
	var result bson.M
	err := r.Collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *QuestionRepository) Update(filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	return r.Collection.UpdateOne(context.TODO(), filter, bson.M{"$set": update})
}

func (r *QuestionRepository) Delete(filter bson.M) (*mongo.DeleteResult, error) {
	return r.Collection.DeleteOne(context.TODO(), filter)
}
