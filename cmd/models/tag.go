package models

type Tag struct {
	ID     uint   `json:"id" gorm:"primary_key;auto_increment;notNull"`
	Title string `json:"title" gorm:"notNull"`
	Slug string `json:"slug" gorm:"notNull" validate:"omitempty"`
	Content string `json:"content"`
}