package models

type Test struct {
	ID              string   `json:"_id"`
	EmailID         string   `json:"email_id"`
	TestName        string   `json:"test_name"`
	Descript        string   `json:"descript"`
	ClassID         []any    `json:"class_ids"`
	QuestionIDs     []any    `json:"question_ids"`
	AllowedUsers    []string `json:"allowed_users"`
	IsTest          bool     `json:"is_test"`
	Random          bool     `json:"random"`
	StartTime       string   `json:"start_time"`
	EndTime         string   `json:"end_time"`
	DurationMinutes int      `json:"duration_minutes"`
	Tags            []string `json:"tags"`
	CreatedAt       any      `json:"created_at"`
	UpdatedAt       any      `json:"updated_at"`
}

type TestId struct {
	ID      string `json:"_id"`
	EmailID string `json:"email_id"`
}
