package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpd "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpd.SearchClient
}

func NewSearchClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: searchpd.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)
	
	response, err := c.client.Search(ctx, &searchpd.SearchRequest{
		Limit:  int64(limit),
		Phrase: phrase,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return nil, core.ErrBadArguments
		}
		return nil, err
	}
	comics := make([]core.Comics, len(response.Comics))
	for i, elem := range response.Comics {
		comics[i] = core.Comics{
			ID:  int(elem.Id),
			Title: elem.Title,
			URL: elem.Url,
		}
	}

	return comics, nil
}

func (c Client) ISearch(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)

	response, err := c.client.ISearch(ctx, &searchpd.SearchRequest{
		Limit:  int64(limit),
		Phrase: phrase,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return nil, core.ErrBadArguments
		}
		return nil, err
	}
	comics := make([]core.Comics, len(response.Comics))
	for i, elem := range response.Comics {
		comics[i] = core.Comics{
			ID:  int(elem.Id),
			Title: elem.Title,
			URL: elem.Url,
		}
	}

	return comics, nil
}

func (c Client) GetByID(ctx context.Context, id int) (core.Comics, error) {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)

	comics, err := c.client.GetById(ctx, &searchpd.ComicsRequest{
		Id: int64(id),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return core.Comics{}, core.ErrNotFound
		}
		return core.Comics{}, err
	}

	return core.Comics{
		ID: int(comics.Id),
		Title: comics.Title,
		URL: comics.Url,
		Description: comics.Description,
		ImgDescription: comics.Imgdescription,
	}, nil
}
