package core

import "context"

type Normalizer interface {
	Pinger
	Norm(context.Context, string) ([]string, error)
}

type Pinger interface {
	Ping(context.Context) error
}

type Updater interface {
	Pinger
	Update(context.Context) error
	Stats(context.Context) (UpdateStats, error)
	Status(context.Context) (UpdateStatus, error)
	Drop(context.Context) error
}

type Searcher interface {
	Pinger
	Search(context.Context, string, int) ([]Comics, error)
	ISearch(context.Context, string, int) ([]Comics, error)
	GetByID(context.Context, int) (Comics, error)
}

type AAA interface {
	Pinger
	Register(context.Context, string, string, string, UserRole) error
	Login(context.Context, string, string) (string, error)
	Verify(context.Context, string) (User, error)
}

type Profile interface {
	LikeComics(context.Context, string, int) error
	DiselikeComics(context.Context, string, int) error
	LikesComics(context.Context, string) ([]Comics, error)
}
