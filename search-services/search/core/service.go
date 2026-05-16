package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log   *slog.Logger
	db    DB
	words Words
	index Index
	broker NatsBroker
}

func NewService(
	log *slog.Logger, db DB, words Words, index Index, broker NatsBroker,
) (*Service, error) {
	return &Service{
		log:   log,
		db:    db,
		words: words,
		index: index,
		broker: broker,
	}, nil
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) ([]Comics, error) {
	if phrase == "" {
		return s.db.GetPage(ctx, limit)
	}
	
	phraseNormalize, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return []Comics{}, err
	}

	resp, err := s.db.Get(ctx, phraseNormalize, limit)
	if err != nil {
		return []Comics{}, err
	}

	return resp, nil
}

func (s *Service) ISearch(ctx context.Context, phrase string, limit int) ([]Comics, error) {
	if phrase == "" {
		return s.db.GetPage(ctx, limit)
	}

	phraseNormalize, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return []Comics{}, err
	}

	resp := s.index.Gets(phraseNormalize)

	respDB, err := s.db.Search(ctx, resp, limit)
	if err != nil {
		return []Comics{}, err
	}

	return respDB, nil
}

func (s *Service) GetByID(ctx context.Context, id int) (Comics, error) {
	respDB, err := s.db.GetByID(ctx, id)
	if err != nil {
		return Comics{}, err
	}

	return respDB, nil
}

func (s *Service) BuildIndex(ctx context.Context) error {
	comics, err := s.db.GetAll(ctx)
	if err != nil {
		return err
	}
	newIndex := make(map[string][]int)
	unique := make(map[int]struct{})
	for _, paper := range comics {
		unique[paper.ID] = struct{}{}
		newIndex[paper.Word] = append(newIndex[paper.Word], paper.ID)
	}
	s.index.Rebuild(newIndex, len(unique))
	return nil
}

func (s *Service) HandlerOptions(ctx context.Context, event EventDB) {
	switch event {
	case EventDBUpdated:
		if err := s.BuildIndex(ctx); err != nil {
			s.log.Error("failed to rebuild index", "error", err)
		}
	case EventDBDrop:
		s.index.Drop()
	default:
		s.log.Error("unknow event from nats", "msg", event)
	}
}
