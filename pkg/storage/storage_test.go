package storage_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/leonideliseev/ozonTestTask/pkg/model"
	"github.com/leonideliseev/ozonTestTask/pkg/storage"
	"github.com/leonideliseev/ozonTestTask/pkg/storage/in_memory"
	"github.com/leonideliseev/ozonTestTask/pkg/storage/postgresql"
	u "github.com/leonideliseev/ozonTestTask/pkg/storage/utils"
)

func TestStorage(t *testing.T) {
	if _, err := os.Stat(".env"); err == nil {
        err = godotenv.Load()
        if err != nil {
            log.Fatalf("Error loading .env file")
        }
		fmt.Println(".env file does loaded succefully")
    } else {
		fmt.Println(".env file does not exist")
	}
	
	connectionString := getEnv("DATABASE_URL", "postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable")

	pgStorage, err := postgresql.NewPostgreStore(connectionString)
	if err != nil {
        t.Fatalf("failed to create PostgresqlStorage: %v", err)
    }

	pgStorage.DB.LogMode(true)
	pgStorage.DB.AutoMigrate(&smodel.Post{}, &smodel.User{}, &smodel.Comment{})

	// тестирует сразу и бд и in-memory
	storages := []struct {
        name    string
        storage storage.Storage
    }{
        {"PostgresqlStorage", pgStorage},
        {"InMemoryStorage", memory.NewInMemoryStore()},
    }

	// корректные данные для создания
	user := u.GetCleanUser()
	post := u.GetCleanPost()
	comm := u.GetCleanComment()
	// будут запоминаться id корректных данных
	var userId, postId, commId uint

	for _, s := range storages {
		t.Run(s.name, func (t *testing.T) {
			t.Run("CreatePost", func(t *testing.T) {
				t.Run("SuccessfulCreatePost", func (t *testing.T) {
					okUser, err := s.storage.CreateUser(user)
					if err != nil {
						t.Errorf("Error create user: %s", err.Error())
					}
					userId = okUser.ID

					post := post
					post.UserId = userId
					okPost, err := s.storage.CreatePost(post)
					if err != nil {
						t.Errorf("Error create post: %s", err.Error())
					}
					postId = okPost.ID
				})

				t.Run("CreatePostWithWrongUserId", func(t *testing.T) {
					post := post
					post.UserId = userId + 1
	
					if _, err := s.storage.CreatePost(post); err.Error() != u.ErrorUserId(post.UserId) {
						t.Error(
							"expected", u.ErrorUserId(post.UserId),
							"got", err.Error(),
						)
					}
				})
			})

			// уже созданы юзер и пост
			t.Run("CreateComment", func(t *testing.T) {
				comm := comm
				comm.UserId = userId
				comm.PostId = postId

				t.Run("SuccessfulCreateComm", func(t *testing.T) {
					okComm, err := s.storage.CreateComment(comm)
					if err != nil {
						t.Errorf("Error create comm: %s", err.Error())
					}
					commId = okComm.ID
				})

				t.Run("CreateCommWithWrongUserId", func(t *testing.T) {
					comm := comm
					comm.UserId = userId + 1

					if _, err := s.storage.CreateComment(comm); err.Error() != u.ErrorUserId(comm.UserId) {
						t.Error(
							"expected", u.ErrorUserId(comm.UserId),
							"got", err.Error(),
						)
					}
				})

				t.Run("CreateCommWithWrongUPostId", func(t *testing.T) {
					comm := comm
					comm.PostId = postId + 1

					if _, err := s.storage.CreateComment(comm); err.Error() != u.ErrorPostId(comm.PostId) {
						t.Error(
							"expected", u.ErrorPostId(comm.PostId),
							"got", err.Error(),
						)
					}
				})

				t.Run("CreateCommForPostWithDisabledComment", func(t *testing.T) {
					post := post
					post.UserId = userId
					post.CommentsEnabled = false
					disablePost, err := s.storage.CreatePost(post)
					if err != nil {
						t.Errorf("Error create post: %s", err.Error())
					}
					
					comm := comm
					comm.PostId = disablePost.ID

					if _, err := s.storage.CreateComment(comm); err.Error() != u.ErrorCommDisable() {
						t.Error(
							"expected", u.ErrorCommDisable(),
							"got", err.Error(),
						)
					}
				})
			})

			// уже созданы два поста (enable/disable), юзер, коммент
			t.Run("CreateReply", func (t *testing.T) {
				reply := comm
				reply.UserId = userId
				reply.PostId = postId
				reply.ParentId = &commId

				t.Run("SuccessfulCreateReply", func(t *testing.T) {
					if _, err := s.storage.CreateComment(reply); err != nil {
						t.Errorf("Error create reply: %s", err.Error())
					}
				})

				// проверять создание reply при неправильном UserId/PostId нет смысла,
				// так как изменение поля ParentId не влияет на эти проверки
				t.Run("CreateReplyWithWrongParentId", func(t *testing.T) {
					reply := reply
					parId := uint(commId + 2) // +1 получил reply
					reply.ParentId = &parId

					if _, err := s.storage.CreateComment(reply); err.Error() != u.ErrorParentIdForReply(parId) {
						t.Error(
							"expected", u.ErrorParentIdForReply(parId),
							"got", err.Error(),
						)
					}
				})

				t.Run("CreateReplyWithMismatchPostId", func(t *testing.T) {
					// создаём второй пост с включёнными комментариями,
					// чтобы указать его как PostId для reply, у которого ответ идёт
					// на комментарий с другим PostId
					if _, err := s.storage.CreatePost(post); err != nil {
						t.Errorf("Error create post: %s", err.Error())
					}

					reply := reply
					reply.PostId = postId + 2 // это новый созданный пост

					if _, err := s.storage.CreateComment(reply); err.Error() != u.ErorrMismatchPostId(reply.PostId, comm.PostId) {
						t.Error(
							"expected", u.ErorrMismatchPostId(reply.PostId, comm.PostId),
							"got", err.Error(),
						)
					}
				})
			})
		})
	}
}

// получение значения из окружения
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
		fmt.Printf("%s set from env\n", key)
        return value
    }
	fmt.Printf("%s set default\n", key)
    return defaultValue
}
