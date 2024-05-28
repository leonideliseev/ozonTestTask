package utils

import (
	"fmt"
	
	"github.com/leonideliseev/ozonTestTask/pkg/model"
)
// функции для получения "чистых" данных (используется для тестов)
func GetCleanUser() smodel.CreateUser {
	return smodel.CreateUser{
		Username: "qwerty",
	}
}

func GetCleanPost() smodel.CreatePost {
	return smodel.CreatePost{
		Title: "TestTitle",
		Content: "TestContent",
		UserId: 1,
		CommentsEnabled: true,
	}
}

func GetCleanComment() smodel.CreateComment {
	return smodel.CreateComment{
		PostId: 1,
		ParentId: nil,
		UserId: 1,
		Content: "Test Content",
	}
}

// функции для получения ошибок
func ErrorUserId(id uint) string {
	return fmt.Sprintf("author with id = %d not found", id)
}

func ErrorPostId(id uint) string {
	return fmt.Sprintf("post with id = %d not found", id)
}

func ErrorCommId(id uint) string {
	return fmt.Sprintf("comment with id = %d not found", id)
}

func ErrorCommDisable() string {
	return "comment not enable for this post"
}

func ErrorParentIdForReply(id uint) string {
	return fmt.Sprintf("comment with id = %d for reply not found", id)
}

func ErorrMismatchPostId(replyPostId, commPostId uint) string {
	return fmt.Sprintf("reply post id = %d doesn't match the comment post id = %d being replied to", replyPostId, commPostId)
}
