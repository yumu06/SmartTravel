package TravelModel

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID      uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID  uint64    `gorm:"not null;index:idx_post_user_created,priority:1"`
	Title   string    `gorm:"type:varchar(30); not null"`
	HeadImg string    `gorm:"type:text"`
	Content string    `gorm:"type:text; not null"`

	ViewCount uint64 `gorm:"not null;default:0;index:idx_post_recommend,priority:2,sort:desc"`
	LikeCount uint64 `gorm:"not null;default:0;index:idx_post_recommend,priority:1,sort:desc"`

	CreatedAt CustomTime `gorm:"index:idx_post_user_created,priority:2;index:idx_post_recommend,priority:3,sort:desc"`
	UpdatedAt CustomTime
}

func (post *Post) BeforeCreate(db *gorm.DB) error {
	db.Statement.SetColumn("id", uuid.NewV4())
	return nil
}

type PostComment struct {
	ID      uuid.UUID `gorm:"type:char(36);primary_key"`
	PostID  uuid.UUID `gorm:"type:char(36);not null;index:idx_comment_post_created,priority:1"`
	UserID  uint64    `gorm:"not null;index"`
	Content string    `gorm:"type:text;not null"`

	CreatedAt CustomTime `gorm:"index:idx_comment_post_created,priority:2"`
	UpdatedAt CustomTime
}

func (c *PostComment) BeforeCreate(db *gorm.DB) error {
	db.Statement.SetColumn("id", uuid.NewV4())
	return nil
}

type Notice struct {
	ID       uuid.UUID `gorm:"type:char(36);primary_key"`
	Title    string    `gorm:"type:varchar(128);not null"`
	Content  string    `gorm:"type:text;not null"`
	ImageURL string    `gorm:"type:text"`
	LinkURL  string    `gorm:"type:varchar(256)"`

	CreatedAt CustomTime
	UpdatedAt CustomTime
}

func (n *Notice) BeforeCreate(db *gorm.DB) error {
	db.Statement.SetColumn("id", uuid.NewV4())
	return nil
}
