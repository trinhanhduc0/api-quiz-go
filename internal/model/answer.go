package models

import (
	"time"
)

type TestAnswer struct {
	ID                 string           `json:"_id,omitempty"`
	TestId             any              `json:"test_id,omitempty"`
	EmailID            string           `json:"email_id,omitempty"`
	Email              string           `json:"email,omitempty"`
	ListQuestionAnswer []QuestionAnswer `json:"question_answer,omitempty"`
	CreatedAt          time.Time        `json:"created_at,omitempty"`
}

type QuestionAnswer struct {
	QuestionID      any              `json:"question_id,omitempty"`
	FillInTheBlanks []FillInTheBlank `json:"fill_in_the_blank,omitempty"`
	Options         []Option         `json:"options,omitempty"`
}
