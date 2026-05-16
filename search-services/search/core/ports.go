package core

import "context"

type Search interface {
	Search(ctx context.Context, words string, limit int) ([]Comics, error)
	ISearch(ctx context.Context, words string, limit int) ([]Comics, error)
	GetByID(ctx context.Context, id int) (Comics, error)
}

type DB interface {
	Search(ctx context.Context, words []int, limit int) ([]Comics, error)
	Get(ctx context.Context, words []string, limit int) ([]Comics, error)
	GetByID(ctx context.Context, id int) (Comics, error)
	GetPage(ctx context.Context, limit int) ([]Comics, error)
	GetAll(ctx context.Context) ([]ComicsFWords, error)
}

type Index interface {
	Set(word string, id int)
	Gets(words []string) []int
	Drop()
	SetTotal(n int)
	Rebuild(newIndex map[string][]int, total int)
}

type Rebuilder interface {
	BuildIndex(ctx context.Context) error
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Initiator interface {
	Status()
}

type NatsBroker interface {
	Subscribe(ctx context.Context, handler func(context.Context, EventDB)) error
}
