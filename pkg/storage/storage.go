package storage

import (
	"github.com/leonideliseev/ozonTestTask/pkg/model"
)

// интерфейс для памяти
type Storage interface {
	CreatePost(p smodel.CreatePost) (*smodel.Post, error)
	CreateComment(c smodel.CreateComment) (*smodel.Comment, error)
	CreateUser(u smodel.CreateUser) (*smodel.User, error)
	GetPosts() ([]*smodel.Post, error)
	GetPost(id uint) (*smodel.Post, error)
	GetComments(id uint) ([]*smodel.Comment, error)
}