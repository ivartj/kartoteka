package entities

import (
	"github.com/google/uuid"
)

type ID []byte

func NewID() ID {
	id := uuid.New()
	return ID(id[:])
}

type UserID ID

type User struct {
	ID              UserID `sqlname:"user_id"`
	Username        string `sqlname:"username"`
	Email           string `sqlname:"email"`
	EmailUnverified string `sqlname:"email_unverified"`
	PasswordHash    string `sqlname:"password_hash"`
}

type Language struct {
	Code       string `sqlname:"language_code"`
	NativeName string `sqlname:"native_name"`
}

type ImageID ID

type Image struct {
	ID             ImageID `sqlname:"image_id"`
	MimeType       string  `sqlname:"mime_type"`
	License        string  `sqlname:"license"`
	Attribution    string  `sqlname:"attribution"`
	AttributionUrl string  `sqlname:"attribution_url"`
}

type WordID ID

type Word struct {
	ID           WordID `sqlname:"word_id"`
	Word         string `sqlname:"word"`
	LanguageCode string `sqlname:"language_code"`
	UserID       UserID `sqlname:"user_id"`
	ImageID      UserID `sqlname:"image_id"`
	Notes        string `sqlname:"notes"`

	Translations []*WordTranslation
	Tags         []string
	UserUsername string
}

type WordTranslation struct {
	WordID       WordID `sqlname:"word_id"`
	LanguageCode string `sqlname:"language_code"`
	Translation  string `sqlname:"translation"`
}

type WordTag struct {
	WordID WordID `sqlname:"word_id"`
	Tag    string `sqlname:"tag"`
}
