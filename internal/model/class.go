package models

import (
	"time"
)

type Class struct {
	ID            string    `json:"_id"`
	ClassName     string    `json:"class_name"`
	AuthorMail    string    `json:"author_mail"`
	TestID        []any     `json:"test_id"`
	EmailID       string    `json:"email_id"`
	StudentAccept []string  `json:"students_accept"`
	StudentsWait  []string  `json:"students_wait"`
	IsPublic      bool      `json:"is_public"`
	UpdatedAt     time.Time `json:"updated_at"`
	Tags          []string  `json:"tags"`
}
