package models

// Course represents one learning course in the system.
type Course struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Level       string `json:"level"`
	Duration    int    `json:"duration"`
	Completed   bool   `json:"completed"`
}
