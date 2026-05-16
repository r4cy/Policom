package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

func NewServer(service core.Updater) *Server {
	return &Server{service: service}
}

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	nwStatus := s.service.Status(ctx)
	var protoStatus updatepb.Status

	switch nwStatus {
	case core.StatusIdle:
		protoStatus = updatepb.Status_STATUS_IDLE
	case core.StatusRunning:
		protoStatus = updatepb.Status_STATUS_RUNNING
	default:
		protoStatus = updatepb.Status_STATUS_UNSPECIFIED
	}

	return &updatepb.StatusReply{
		Status: protoStatus,
	}, nil
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if s.service.Status(ctx) == core.StatusRunning {
		return &emptypb.Empty{}, status.Error(codes.AlreadyExists, "update already running")
	}

	upCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	go func() {
		defer cancel()
		//nolint:errcheck
		s.service.Update(upCtx)
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	stats, err := s.service.Stats(ctx)
	if err != nil {
		return &updatepb.StatsReply{}, err
	}
	return &updatepb.StatsReply{
		WordsTotal:    int64(stats.DBStats.WordsTotal),
		WordsUnique:   int64(stats.DBStats.WordsUnique),
		ComicsTotal:   int64(stats.ComicsTotal),
		ComicsFetched: int64(stats.DBStats.ComicsFetched),
	}, nil
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Drop(ctx); err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}
