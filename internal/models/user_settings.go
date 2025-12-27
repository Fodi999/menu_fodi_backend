package models

import "errors"

// Language types
type Language string

const (
	LanguagePolish  Language = "pl"
	LanguageEnglish Language = "en"
	LanguageRussian Language = "ru"
)

// TimeFormat types
type TimeFormat string

const (
	TimeFormat12h TimeFormat = "12h"
	TimeFormat24h TimeFormat = "24h"
)

// Units types
type Units string

const (
	UnitsMetric  Units = "metric"
	UnitsKitchen Units = "kitchen"
)

// AIStyle types
type AIStyle string

const (
	AIStyleMentor AIStyle = "mentor"
	AIStyleDirect AIStyle = "direct"
)

// UserSettings represents user preferences
// Stored as JSONB in database
type UserSettings struct {
	Language   Language   `json:"language"`   // pl | en | ru
	TimeFormat TimeFormat `json:"timeFormat"` // 12h | 24h
	Units      Units      `json:"units"`      // metric | kitchen
	AIStyle    AIStyle    `json:"aiStyle"`    // mentor | direct
}

// DefaultUserSettings returns default settings (Polish locale)
func DefaultUserSettings() UserSettings {
	return UserSettings{
		Language:   LanguagePolish,
		TimeFormat: TimeFormat24h,
		Units:      UnitsMetric,
		AIStyle:    AIStyleMentor,
	}
}

// UpdateSettingsRequest for PATCH /api/settings
// All fields are optional (pointer types allow partial updates)
type UpdateSettingsRequest struct {
	Language   *Language   `json:"language,omitempty"`
	TimeFormat *TimeFormat `json:"timeFormat,omitempty"`
	Units      *Units      `json:"units,omitempty"`
	AIStyle    *AIStyle    `json:"aiStyle,omitempty"`
}

// Validate checks if settings values are valid
func (s *UserSettings) Validate() error {
	// Language validation
	if s.Language != LanguagePolish && s.Language != LanguageEnglish && s.Language != LanguageRussian {
		return ErrInvalidLanguage
	}

	// TimeFormat validation
	if s.TimeFormat != TimeFormat12h && s.TimeFormat != TimeFormat24h {
		return ErrInvalidTimeFormat
	}

	// Units validation
	if s.Units != UnitsMetric && s.Units != UnitsKitchen {
		return ErrInvalidUnits
	}

	// AIStyle validation
	if s.AIStyle != AIStyleMentor && s.AIStyle != AIStyleDirect {
		return ErrInvalidAIStyle
	}

	return nil
}

// Settings validation errors
var (
	ErrInvalidLanguage   = errors.New("invalid language, must be: pl, en, ru")
	ErrInvalidTimeFormat = errors.New("invalid timeFormat, must be: 12h, 24h")
	ErrInvalidUnits      = errors.New("invalid units, must be: metric, kitchen")
	ErrInvalidAIStyle    = errors.New("invalid aiStyle, must be: mentor, direct")
)
