package aaa

import (
	"context"

	"yadro.com/course/api/core"
)

func (a *AAA) LikeComics(ctx context.Context, user_id string, comics_id int) error {
	return a.db.LikeComics(ctx, user_id, comics_id)
}

func (a *AAA) DiselikeComics(ctx context.Context, user_id string, comics_id int) error {
	return a.db.DiselikeComics(ctx, user_id, comics_id)
}

func (a *AAA) LikesComics(ctx context.Context, user_id string) ([]core.Comics, error) {
	comics, err := a.db.LikesComics(ctx, user_id)
	if err != nil {
		return nil, err
	}
	return comics, nil
}