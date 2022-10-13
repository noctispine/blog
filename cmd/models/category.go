package models

type Category struct {
	ID     int64   `json:"id" gorm:"primary_key;auto_increment;notNull"`
	ParentID *int64 `json:"parentId" gorm:"column:parent_id" validate:"omitempty"`
	Title string `json:"title" gorm:"notNull" validate:"required,min=1,max=75"`
	Slug string `json:"slug" gorm:"unique;notNull" validate:"omitempty"`
	Content string `json:"content" validate:"required"`
	
}


type PostCategory struct {
	PostID int64 `json:"postId" gorm:"column:post_id;notNull"`
	CategoryID int64 `json:"categoryID" gorm:"column:category_id;notNull"`
}