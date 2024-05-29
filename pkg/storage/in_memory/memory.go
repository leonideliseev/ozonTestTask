package memory

import (
	"errors"

	"sync"

	"github.com/leonideliseev/ozonTestTask/pkg/model"
	u "github.com/leonideliseev/ozonTestTask/pkg/storage/utils"
)

type MemoryStorage struct {
	posts   map[uint]smodel.Post
    users   map[uint]smodel.User
    comments map[uint]smodel.Comment

	commReply map[int][]int
	// commReply является вспомогательной структурой.
	// Она создана для того, чтобы получать комментарии, которые относятся к посту/другому комментарию по id
	// Для однозначного соответствия, id постов в эту мапу заносятся как -id (отрицательные)
	// Иначе происходили бы коллизии между id постов и комментариев

    mu      sync.RWMutex
}

func NewInMemoryStore() *MemoryStorage {
	return &MemoryStorage{
		posts:    make(map[uint]smodel.Post),
        users:    make(map[uint]smodel.User),
        comments: make(map[uint]smodel.Comment),
		commReply: make(map[int][]int),
	}
}

func (m *MemoryStorage) CreatePost(p smodel.CreatePost) (*smodel.Post, error) {
	m.mu.Lock()
    defer m.mu.Unlock()

	user, ok := m.users[p.UserId]
	if !ok {
		return nil, errors.New(u.ErrorUserId(p.UserId))
	}

	id := uint(len(m.posts)) + 1 // +1 чтобы id совпадал с бд и не было коллизий в функции getComm

	post := smodel.Post{
		ID: id,
		Title: p.Title,
		Content: p.Content,
		UserID: p.UserId,
		User: user,
		CommentsEnabled: p.CommentsEnabled,
		CommPage: &smodel.CommPage{
			Comms: []*smodel.Comment{},
			TotalCount: 0,
		},
	}

    m.posts[id] = post

    return &post, nil
}

func (m *MemoryStorage) CreateComment(c smodel.CreateComment) (*smodel.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// проверка существования автора
	user, ok := m.users[c.UserId]
	if !ok {
		return nil, errors.New(u.ErrorUserId(c.UserId))
	}

	// проверка существования поста
	post, exist := m.posts[c.PostId]
	if !exist {
		return nil, errors.New(u.ErrorPostId(c.PostId))
	}

	// проверка что можно оставлять комментарии
	if !post.CommentsEnabled {
		return nil, errors.New(u.ErrorCommDisable())
	}

	// если ответ на другой комментарий
	if c.ParentId != nil {
		parentComm, ok := m.comments[*c.ParentId]
		// проверка существования родительского поста
		if !ok {
			return nil, errors.New(u.ErrorParentIdForReply(*c.ParentId))
		}

		// проверка чтобы ответ на комментарий был под тем же постом
		if c.PostId != parentComm.PostID {
			return nil, errors.New(u.ErorrMismatchPostId(c.PostId, parentComm.PostID))
		}
	}

	id := uint(len(m.comments)) + 1 // чтобы совпадало с бд

	comment := smodel.Comment{
		ID: id,
		PostID: c.PostId,
		ParentID: c.ParentId,
		UserID: c.UserId,
		User: user,
		Content: c.Content,
	}

	m.comments[id] = comment

	// добавление в commReply
	if comment.ParentID != nil { // если комментарий на комментарий
		//if _, exist := m.commReply[int(*comment.ParentID)]; exist {
	    m.commReply[int(*comment.ParentID)] = append(m.commReply[int(*comment.ParentID)], int(id))
		/*} else {
			m.commReply[int(*comment.ParentID)] = []int{int(id)}
		}*/
	} else { // если комментарий под пост
		m.commReply[-int(comment.PostID)] = append(m.commReply[-int(comment.PostID)], int(id))
	}

	return &comment, nil
}

func (m *MemoryStorage) CreateUser(u smodel.CreateUser) (*smodel.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	id := uint(len(m.users)) + 1 // чтобы совпадало с бд
	
	user := smodel.User{
		ID: id,
		Username: u.Username,
	}

	m.users[id] = user

	return &user, nil
}

func (m *MemoryStorage) GetPosts(limit, offset int) (*smodel.PostPage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalCount := len(m.posts)

	posts := make([]*smodel.Post, 0, limit)
	for postId, post := range m.posts {
		if int(postId) > offset && int(postId) <= offset + limit {
			posts = append(posts, &post)
		}
	}

	return &smodel.PostPage{
		Posts: posts,
		TotalCount: totalCount,
	}, nil
}

func (m *MemoryStorage) GetPost(limit, offset int, id uint) (*smodel.Post, error) {
	m.mu.RLock()
    defer m.mu.RUnlock()

	// проверка существования поста
    post, ok := m.posts[id]
    if !ok {
        return nil, errors.New(u.ErrorPostId(id))
    }

	// получение комментариев к посту
	//comms := m.getComments(limit, offset, -int(id), 0)

	post.CommPage = m.getComments(limit, offset, -int(id), 0)

    return &post, nil
}

func (m *MemoryStorage) GetComments(limit, offset int, id uint) (*smodel.Comment, error) {
	m.mu.RLock()
    defer m.mu.RUnlock()

	// проверка существования комментария
	comm, ok := m.comments[id]
	if !ok {
		return nil, errors.New(u.ErrorCommId(id))
	}

	// получение комментариев
	// начинаем с глубины 1, так как уже есть сам комментарий
	comm.ReplyPage = m.getComments(limit, offset, int(id), 1)
	//comms := []*smodel.Comment{&comm}

	return &comm, nil
}

// рекурсивно получает комментарии
func (m *MemoryStorage) getComments(limit, offset, id, depth int) *smodel.CommPage {
	m.mu.RLock()
    defer m.mu.RUnlock()

	commPage := smodel.CommPage{
		Comms: make([]*smodel.Comment, 0),
		TotalCount: 0,
	}

	comms := make([]*smodel.Comment, 0)

	if depth > 4 {
		return &commPage
	}

	level, ok := m.commReply[id]
	if !ok {
		return &commPage
	}

	totalCount := len(level)
	level = level[offset:offset+limit]

	for _, lv := range level {
		comm := m.comments[uint(lv)]
		comm.ReplyPage = m.getComments(limit, offset, lv, depth + 1)
		comms = append(comms, &comm)
	}

	commPage.Comms = comms
	commPage.TotalCount = totalCount

	return &commPage
}
