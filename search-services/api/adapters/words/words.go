package words

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
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewWordClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: wordspb.NewWordsClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)
	
	response, err := c.client.Norm(ctx, &wordspb.WordsRequest{Phrase: phrase})
	if err != nil {
		switch status.Code(err) {
		case codes.ResourceExhausted:
			c.log.Error(
				"normalize function give error because message > 10 KiB",
				"err", err,
			)
			return []string{}, core.ErrResourceExhausted
		case codes.DeadlineExceeded:
			c.log.Error(
				"normalize function give error because deadline exceed",
				"err", err,
			)
			return []string{}, core.ErrDeadlineExceeded
		default:
			c.log.Error(
				"normalize function give unknow error",
				"err", err,
			)
			return []string{}, core.ErrUnknow
		}
	}
	return response.Words, nil
}
