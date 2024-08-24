package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Question represents the main question structure.
type Question struct {
	ID              string           `json:"_id"`
	FillInTheBlanks []FillInTheBlank `json:"fill_in_the_blank"`
	Metadata        Metadata         `json:"metadata"`
	Options         []Option         `json:"options"`
	QuestionContent QuestionContent  `json:"question_content"`
	Suggestion      string           `json:"suggestion"`
	Tags            []string         `json:"tags"`
	Type            string           `json:"type"`
}

// FillInTheBlank represents a fill-in-the-blank part of the question.
type FillInTheBlank struct {
	TextBefore    string `json:"text_before"`
	Blank         string `json:"blank"`
	CorrectAnswer string `json:"correct_answer"`
	TextAfter     string `json:"text_after"`
}

// Metadata represents the metadata for the question.
type Metadata struct {
	Author string `json:"author"`
}

// Option represents each option in the question.
type Option struct {
	ID        primitive.ObjectID `json:"id"`
	Text      string             `json:"text"`
	ImageURL  string             `json:"image_url"`
	IsCorrect bool               `json:"iscorrect"`
	Match     string             `json:"match"`
	Order     int                `json:"order"`
}

// QuestionContent represents the content of the question.
type QuestionContent struct {
	Text     string `json:"text"`
	ImageURL string `json:"image_url"`
	VideoURL string `json:"video_url"`
	AudioURL string `json:"audio_url"`
}
