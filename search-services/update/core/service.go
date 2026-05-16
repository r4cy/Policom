package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	status      ServiceStatus
	mu          sync.Mutex
	progress    atomic.Bool
	broker      Nats
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, broker Nats, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		status:      StatusIdle,
		broker:      broker,
	}, nil
}

func (s *Service) Update(ctx context.Context) (err error) {
	if ok := s.mu.TryLock(); !ok {
		return ErrAlreadyRunning
	}
	defer s.mu.Unlock()

	s.progress.Store(true)
	defer s.progress.Store(false)

	lastComics, err := s.xkcd.LastID(ctx)
	if err != nil {
		return err
	}
	comicsExists, err := s.db.IDs(ctx)
	if err != nil {
		return err
	}

	//nolint:errcheck
	defer s.broker.Publish(ctx, "XKCD DB has been updated")

	comicsExistsMAP := make(map[int]struct{}, len(comicsExists))
	missingComics := make([]int, 0)

	for _, elem := range comicsExists {
		comicsExistsMAP[elem] = struct{}{}
	}

	for i := 1; i <= lastComics; i++ {
		if _, ok := comicsExistsMAP[i]; !ok {
			missingComics = append(missingComics, i)
		}
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for id := range jobs {
				comics, err := s.xkcd.Get(ctx, id)
				if err != nil {
					continue
				}

				normalize, err := s.words.Norm(ctx, fmt.Sprintf("%s %s %s", comics.Description, comics.Title, comics.ImgDescription))
				if err != nil {
					continue
				}

				if err := s.db.Add(ctx, Comics{
					ID:             comics.ID,
					URL:            comics.URL,
					Words:          normalize,
					Title:          comics.Title,
					Description:    comics.Description,
					ImgDescription: comics.ImgDescription,
				}); err != nil {
					continue
				}
			}
		}()
	}

	for _, elem := range missingComics {
		jobs <- elem
	}

	close(jobs)
	wg.Wait()

	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: comicsTotal - 1, // Захардкодил -1 для комикса 404
	}, nil

}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.progress.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.db.Drop(ctx); err != nil {
		return err
	}
	if err := s.broker.Publish(ctx, "XKCD DB has been drop"); err != nil {
		return err
	}
	return nil
}
