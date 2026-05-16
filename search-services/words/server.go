package main

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	norm "yadro.com/course/words/words"
)

type server struct {
	wordspb.UnimplementedWordsServer
	log *slog.Logger
}

func NewServer(log *slog.Logger) *server {
	return &server{
		log: log,
	}
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if in.GetPhrase() == "" {
		s.log.Error("failed in Norm method", "err", "you can't send nil string")
		return &wordspb.WordsReply{
				Words: []string{},
			},
			nil
	}
	normalize := norm.NormalizeTheWords(in.Phrase)
	return &wordspb.WordsReply{
		Words: normalize,
	}, nil
}
