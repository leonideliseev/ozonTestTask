package graph

import (
	"sync"

	"github.com/leonideliseev/ozonTestTask/graph/model"
	"github.com/leonideliseev/ozonTestTask/pkg/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	storage storage.Storage
	subscribers  map[string]chan *model.Comment
	mu           sync.RWMutex
}

func NewResolver(store storage.Storage) *Resolver {
    return &Resolver{
		storage: store,
		subscribers: make(map[string]chan *model.Comment),
	}
}
