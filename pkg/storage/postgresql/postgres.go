package postgresql

import (
	"errors"

	"github.com/leonideliseev/ozonTestTask/pkg/model"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	u "github.com/leonideliseev/ozonTestTask/pkg/storage/utils"
)

type PostgreStorage struct {
	DB *gorm.DB
}

func NewPostgreStore(connectionString string) (*PostgreStorage, error) {
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&smodel.User{}, &smodel.Post{}, &smodel.Comment{})
	return &PostgreStorage{DB: db}, nil
}

func (s *PostgreStorage) CreatePost(p smodel.CreatePost) (*smodel.Post, error) {
	err := s.checkUserExists(p.UserId)
	if err != nil {
		return nil, err
	}

	post := smodel.Post{
		Title: p.Title,
		Content: p.Content,
		UserID: p.UserId,
		CommentsEnabled: p.CommentsEnabled,
	}

	if err := s.DB.Create(&post).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostgreStorage) CreateComment(c smodel.CreateComment) (*smodel.Comment, error) {
	// проверка существования автора
	err := s.checkUserExists(c.UserId)
	if err != nil {
		return nil, err
	}

	// проверка существованя поста и что можно оставлять комментарии
	err = s.checkPost(c.PostId)
	if err != nil {
		return nil, err
	}

	// если ответ на другой комментарий
	if c.ParentId != nil {
		// проверка существования родительского поста и совпадения их id поста
		err = s.checkParentId(c.PostId, *c.ParentId)
		if err != nil {
			return nil, err
		}
	}

	comment := smodel.Comment{
		PostID: c.PostId,
		ParentID: c.ParentId,
		UserID: c.UserId,
		Content: c.Content,
	}

	if err := s.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

func (s *PostgreStorage) CreateUser(u smodel.CreateUser) (*smodel.User, error) {
	user := smodel.User{
		Username: u.Username,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *PostgreStorage) GetPosts(limit, offset int) (*smodel.PostPage, error) {
	var posts []*smodel.Post
	var totalCount int

	if err := s.DB.Model(&smodel.Post{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Preload("User").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, err
	}

	return &smodel.PostPage{
		Posts: posts,
		TotalCount: totalCount,
	}, nil
}

func (s *PostgreStorage) GetPost(limit, offset int, id uint) (*smodel.Post, error) {
	var post smodel.Post

	if err := s.DB.Preload("User").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Where("parent_id IS NULL").Offset(offset).Limit(limit)
	}).Preload("Comments.User").First(&post, id).Error; err != nil {
		// проверка существования поста
		if gorm.IsRecordNotFoundError(err) {
            return nil, errors.New(u.ErrorPostId(id))
        }
		return nil, err
	}	

	// получение комментариев к посту
	// начинаем с глубины 1, так как уже есть ответы на пост
	comms := post.Comments
	for _, comm := range comms {
		subComms, _ := s.getComments(limit, offset, (*comm).ID, 1)
		(*comm).ReplyPage = subComms
	}

	var totalCount int
	if err := s.DB.Model(&smodel.Comment{}).Where("post_id = ? AND parent_id IS NULL", id).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	post.CommPage = &smodel.CommPage{
		Comms: comms,
		TotalCount: totalCount,
	}

	return &post, nil
}

func (s *PostgreStorage) GetComments(limit, offset int, id uint) (*smodel.Comment, error) {
	var comm smodel.Comment
	
	if err := s.DB.Preload("User").First(&comm, id).Error; err != nil {
		// проверка существования комментария
		if gorm.IsRecordNotFoundError(err) {
            return nil, errors.New(u.ErrorCommId(id))
        }
		return nil, err
	}

	// получение комментариев
	// начинаем с глубины 1, так как уже есть сам комментарий
	var err error
	comm.ReplyPage, err = s.getComments(limit, offset, id, 1)
	if err != nil {
		return nil, err
	}

    return &comm, nil
}

// рекурсивно получает комментарии
func (s *PostgreStorage) getComments(limit, offset int, id uint, depth int) (*smodel.CommPage, error) {
	commPage := smodel.CommPage{
		Comms: make([]*smodel.Comment, 0),
		TotalCount: 0,
	}

	var comms []*smodel.Comment

	if depth > 4 {
		return &commPage, nil
	}

	if err := s.DB.Preload("User").Where("parent_id = ?", id).Offset(offset).Limit(limit).Find(&comms).Error; err != nil {
		return nil, err
	}

	if len(comms) == 0 {
		return &commPage, nil
	}

	var totalCount int
	if err := s.DB.Model(&smodel.Comment{}).Where("parent_id = ?", id).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	commPage.TotalCount = totalCount
	
	for i := range comms {
		childComments, err := s.getComments(limit, offset, comms[i].ID, depth + 1)
		if err != nil {
			return nil, err
		}
		comms[i].ReplyPage = childComments
	}

	commPage.Comms = comms

	return &commPage, nil
}

func (s *PostgreStorage) checkUserExists(userID uint) error {
    var user smodel.User
    if err := s.DB.First(&user, userID).Error; err != nil {
        if gorm.IsRecordNotFoundError(err) {
            return errors.New(u.ErrorUserId(userID))
        }
        return err
    }
    return nil
}

func (s *PostgreStorage) checkPost(postID uint) error {
    var post smodel.Post

	// проверка существования поста
    if err := s.DB.First(&post, postID).Error; err != nil {
        if gorm.IsRecordNotFoundError(err) {
            return errors.New(u.ErrorPostId(postID))
        }
        return err
    }

	// проверка что можно оставлять комментарии
	if !post.CommentsEnabled {
		return errors.New(u.ErrorCommDisable())
	}

    return nil
}

func (s *PostgreStorage) checkParentId(postId, parentId uint) error {
	var comm smodel.Comment

	// проверка существования родительского поста
	if err := s.DB.First(&comm, parentId).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
            return errors.New(u.ErrorParentIdForReply(parentId))
        }
		return err
	}

	// проверка чтобы ответ на комментарий был под тем же постом
	if postId != comm.PostID {
		return errors.New(u.ErorrMismatchPostId(postId, comm.PostID))
	}

	return nil
}
