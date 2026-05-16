package update

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
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewUpdateClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)
	
	response, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		return core.StatusUpdateUnknown, err
	}
	switch response.Status {
	case updatepb.Status_STATUS_IDLE:
		return core.StatusUpdateIdle, err
	case updatepb.Status_STATUS_RUNNING:
		return core.StatusUpdateRunning, err
	default:
		return core.StatusUpdateUnknown, err
	}
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)

	resp, err := c.client.Stats(ctx, &emptypb.Empty{})
	if err != nil {
		return core.UpdateStats{}, err
	}
	return core.UpdateStats{
		WordsTotal:    int(resp.WordsTotal),
		WordsUnique:   int(resp.WordsUnique),
		ComicsFetched: int(resp.ComicsFetched),
		ComicsTotal:   int(resp.ComicsTotal),
	}, nil

}

func (c Client) Update(ctx context.Context) error {
	trace_id := core.TraceIDFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "trace_id", trace_id)

	_, err := c.client.Update(ctx, &emptypb.Empty{})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return core.ErrAlreadyRunning
		}
		return err
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	_, err := c.client.Drop(ctx, &emptypb.Empty{})
	return err
}
