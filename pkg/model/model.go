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
	ID              uint       `gorm:"primary_key"`
	Title           string     `gorm:"not null"`
	Content         string     `gorm:"not null"`
	UserID          uint       `gorm:"not null"`
	User            User       `gorm:"foreignkey:UserID"`
	CommentsEnabled bool       `gorm:"not null"`
	Comments        []*Comment `gorm:"foreignkey:PostID"`
	CommPage        *CommPage   `gorm:"-"`
}

type Comment struct {
	ID        uint       `gorm:"primary_key"`
	PostID    uint       `gorm:"not null"`
	UserID    uint       `gorm:"not null"`
	User      User       `gorm:"foreignkey:UserID"`
	Content   string     `gorm:"not null"`
	ParentID  *uint
	Replies   []*Comment `gorm:"foreignkey:ParentID"`
	ReplyPage *CommPage   `gorm:"-"`
}

type PostPage struct {
	Posts []*Post
	TotalCount int
}

type CommPage struct {
	Comms []*Comment
	TotalCount int
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
	var totalCount int
	var comments []*model.Comment
	if p.CommPage != nil {
		totalCount = p.CommPage.TotalCount
		comments = make([]*model.Comment, len(p.CommPage.Comms))
		for i, comment := range p.CommPage.Comms {
			comments[i] = comment.ToGraphQL()
		}
	}

	user := p.User.ToGraphQL()

	return &model.Post{
		ID:       strconv.FormatUint(uint64(p.ID), 10),
		Title:    p.Title,
		Content:  p.Content,
		UserID:   strconv.FormatUint(uint64(p.UserID), 10),
		Author:   user,
		CommentsEnabled: p.CommentsEnabled,
		CommPage: &model.CommPage{
			Comments: comments,
			TotalCount: totalCount,
		},
	}
}

func (c *Comment) ToGraphQL() *model.Comment {
	var parentID *string
	if c.ParentID != nil {
		idStr := strconv.FormatUint(uint64(*c.ParentID), 10)
		parentID = &idStr
	}

	var totalCount int
	var replies []*model.Comment
	if c.ReplyPage != nil {
		totalCount = c.ReplyPage.TotalCount
		replies = make([]*model.Comment, len(c.ReplyPage.Comms))
		for i, child := range c.ReplyPage.Comms {
			replies[i] = child.ToGraphQL()
		}
	}

	return &model.Comment{
		ID:       strconv.FormatUint(uint64(c.ID), 10),
		PostID:   strconv.FormatUint(uint64(c.PostID), 10),
		ParentCommentID: parentID,
		UserID:   strconv.FormatUint(uint64(c.UserID), 10),
		Author:     c.User.ToGraphQL(),
		Content:  c.Content,
		ReplyPage: &model.CommPage{
			Comments: replies,
			TotalCount: totalCount,
		},
	}
}

func (u *User) ToGraphQL() *model.User {
	return &model.User{
		ID:       strconv.FormatUint(uint64(u.ID), 10),
		Username: u.Username,
	}
}
