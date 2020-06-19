package entities

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/google/uuid"
)

type ID sql.NullString

func NewID() ID {
	var s sql.NullString
	err := s.Scan(uuid.New().String())
	if err != nil {
		panic(err)
	}
	return ID(s)
}

type UserID ID

func (id *UserID) Scan(val interface{}) error {
	return (*sql.NullString)(id).Scan(val)
}

func (id UserID) Value() (driver.Value, error) {
	return sql.NullString(id).Value()
}

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

func (id *ImageID) Scan(val interface{}) error {
	return (*sql.NullString)(id).Scan(val)
}

func (id ImageID) Value() (driver.Value, error) {
	return sql.NullString(id).Value()
}

type Image struct {
	ID             ImageID `sqlname:"image_id"`
	MimeType       string  `sqlname:"mime_type"`
	License        string  `sqlname:"license"`
	Attribution    string  `sqlname:"attribution"`
	AttributionUrl string  `sqlname:"attribution_url"`
}

type WordID ID

func (id *WordID) Scan(val interface{}) error {
	return (*sql.NullString)(id).Scan(val)
}

func (id WordID) Value() (driver.Value, error) {
	return sql.NullString(id).Value()
}

func (id *WordID) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}
	return id.Scan(str)
}

func (id WordID) Marshal() ([]byte, error) {
	str, err := id.Value()
	if err != nil {
		return nil, err
	}
	return json.Marshal(str.(string))
}

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
	WordID       WordID `sqlname:"word_id" json:"word_id"`
	LanguageCode string `sqlname:"language_code" json:"language_code"`
	Translation  string `sqlname:"translation" json:"translation"`
}

type WordTag struct {
	WordID WordID `sqlname:"word_id"`
	Tag    string `sqlname:"tag"`
}
