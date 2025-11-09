package dto

import "time"

// ContactRequest - запрос контактной формы
type ContactRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject"`
	Message string `json:"message" binding:"required"`
}

// ContactResponse - ответ на отправку формы
type ContactResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ContactSubmission - запись в БД (опционально)
type ContactSubmission struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}
