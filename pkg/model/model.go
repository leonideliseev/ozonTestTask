package smodel
// smodel - storage model
// Здесь содержится описание структур, которыми данные будут хранитсья в памяти
// Также функции для перевода из структуры из памяти в структуру для graphQL

import (
	"strconv"

	"github.com/leonideliseev/ozonTestTask/graph/model"
	_ "github.com/lib/pq"
)

type User struct {
	ID       uint   `gorm:"primary_key"`
	Username string `gorm:"not null"`
}

type Post struct {
	ID       uint     `gorm:"primary_key"`
	Title    string   `gorm:"not null"`
	Content  string   `gorm:"not null"`
	UserID   uint     `gorm:"not null"`
	User     User     `gorm:"foreignkey:UserID"`
	CommentsEnabled bool `gorm:"not null"`
	Comments []*Comment `gorm:"foreignkey:PostID"`
}

type Comment struct {
	ID       uint    `gorm:"primary_key"`
	PostID   uint    `gorm:"not null"`
	UserID   uint    `gorm:"not null"`
	User     User    `gorm:"foreignkey:UserID"`
	Content  string  `gorm:"not null"`
	ParentID *uint
	Replies []*Comment `gorm:"foreignkey:ParentID"`
}

type CreatePost struct {
	Title    string
	Content  string
	UserId   uint
	CommentsEnabled bool
}

type CreateComment struct {
	PostId   uint
	UserId   uint
	Content  string
	ParentId *uint
}

type CreateUser struct {
	Username string
}

func (p *Post) ToGraphQL() *model.Post {
	comments := make([]*model.Comment, len(p.Comments))
	for i, comment := range p.Comments {
		comments[i] = comment.ToGraphQL()
	}
	user := p.User.ToGraphQL()
	return &model.Post{
		ID:       strconv.FormatUint(uint64(p.ID), 10),
		Title:    p.Title,
		Content:  p.Content,
		UserID:   strconv.FormatUint(uint64(p.UserID), 10),
		Author:   user,
		CommentsEnabled: p.CommentsEnabled,
		Comments: comments,
	}
}

func (c *Comment) ToGraphQL() *model.Comment {
	var parentID *string
	if c.ParentID != nil {
		idStr := strconv.FormatUint(uint64(*c.ParentID), 10)
		parentID = &idStr
	}
	children := make([]*model.Comment, len(c.Replies))
	for i, child := range c.Replies {
		children[i] = child.ToGraphQL()
	}
	return &model.Comment{
		ID:       strconv.FormatUint(uint64(c.ID), 10),
		PostID:   strconv.FormatUint(uint64(c.PostID), 10),
		ParentCommentID: parentID,
		UserID:   strconv.FormatUint(uint64(c.UserID), 10),
		Author:     c.User.ToGraphQL(),
		Content:  c.Content,
		Replies: children,
	}
}

func (u *User) ToGraphQL() *model.User {
	return &model.User{
		ID:       strconv.FormatUint(uint64(u.ID), 10),
		Username: u.Username,
	}
}
