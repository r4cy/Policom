package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpd "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

const defaultLimit = 10

func NewServer(service core.Search) *Server {
	return &Server{service: service}
}

type Server struct {
	searchpd.UnimplementedSearchServer
	service core.Search
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Search(ctx context.Context, in *searchpd.SearchRequest) (*searchpd.SearchReply, error) {
	if in.GetLimit() == 0 {
		in.Limit = defaultLimit
	}

	comics, err := s.service.Search(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		return &searchpd.SearchReply{}, status.Error(codes.Internal, err.Error())
	}

	massOfComics := make([]*searchpd.Comics, len(comics))
	for i, elem := range comics {
		massOfComics[i] = &searchpd.Comics{
			Id:    int64(elem.ID),
			Title: elem.Title,
			Url:   elem.URL,
		}
	}

	return &searchpd.SearchReply{
		Comics: massOfComics,
		Total:  int64(len(comics)),
	}, nil
}

func (s *Server) ISearch(ctx context.Context, in *searchpd.SearchRequest) (*searchpd.SearchReply, error) {
	if in.GetLimit() == 0 {
		in.Limit = defaultLimit
	}

	comics, err := s.service.ISearch(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		return &searchpd.SearchReply{}, status.Error(codes.Internal, err.Error())
	}

	massOfComics := make([]*searchpd.Comics, len(comics))
	for i, elem := range comics {
		massOfComics[i] = &searchpd.Comics{
			Id:    int64(elem.ID),
			Title: elem.Title,
			Url:   elem.URL,
		}
	}

	return &searchpd.SearchReply{
		Comics: massOfComics,
		Total:  int64(len(comics)),
	}, nil
}

func (s *Server) GetById(ctx context.Context, in *searchpd.ComicsRequest) (*searchpd.Comics, error) {
	comics, err := s.service.GetByID(ctx, int(in.GetId()))
	if err != nil {
		return &searchpd.Comics{}, status.Error(codes.NotFound, err.Error())
	}

	return &searchpd.Comics{
		Id: int64(comics.ID),
		Title: comics.Title,
		Url: comics.URL,
		Description: comics.Description,
		Imgdescription: comics.ImgDescription,
	}, nil
}
