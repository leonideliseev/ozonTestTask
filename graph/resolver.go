package graph

import (
	"github.com/leonideliseev/ozonTestTask/pkg/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	storage storage.Storage
}

func NewResolver(store storage.Storage) *Resolver {
    return &Resolver{storage: store}
}
