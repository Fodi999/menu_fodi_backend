package service

import "errors"

// Доменные ошибки для правильного HTTP маппинга

var (
	// ErrNotFound - продукт не найден (404)
	ErrNotFound = errors.New("fridge item not found")

	// ErrForbidden - продукт не принадлежит пользователю (403)
	ErrForbidden = errors.New("access denied: item does not belong to user")

	// ErrInvalidInput - некорректные входные данные (400)
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidSource - некорректный source для цены (400)
	ErrInvalidSource = errors.New("invalid price source")
)
