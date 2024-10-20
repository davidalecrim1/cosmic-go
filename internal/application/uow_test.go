package application

import (
	"context"
)

type FakeUoW struct {
	adapters Adapters
}

func NewFakeUnitOfWorkFromRepository(repo Repository) *FakeUoW {
	return &FakeUoW{adapters: Adapters{Repository: repo}}
}

func (u *FakeUoW) Transact(ctx context.Context, txFunc func(_ Adapters) error) error {
	return txFunc(u.adapters)
}
