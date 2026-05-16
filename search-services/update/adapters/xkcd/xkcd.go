package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/update/core"
)

type responseLastID struct {
	Num int `json:"num"`
}

type responseGetComics struct {
	ID             int    `json:"num"`
	URL            string `json:"img"`
	Title          string `json:"title"`
	Description    string `json:"transcript"`
	ImgDescription string `json:"alt"`
	SafeTitle      string `json:"safe_title"`
}

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewXKCDClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

// Скачивание комикса с сайте XKCD с нужным id
func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/%d/info.0.json", c.url, id))
	if err != nil {
		c.log.Error(
			"failed to do get request",
			"url", c.url,
			"id", id,
			"err", err,
		)
		return core.XKCDInfo{}, core.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		c.log.Error(
			"failed to do get request",
			"url", c.url,
			"id", id,
			"statusCode", resp.StatusCode,
		)
		return core.XKCDInfo{}, core.ErrNotFound
	}
	//nolint:errcheck // close error because non-actionable here
	defer resp.Body.Close()

	var response responseGetComics
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.log.Error(
			"failed to decode information from comics",
			"url", c.url,
			"id", id,
			"err", err,
		)
		return core.XKCDInfo{}, core.ErrNotFound
	}

	return core.XKCDInfo{
		ID:             response.ID,
		URL:            response.URL,
		Title:          response.Title,
		Description:    response.Description,
		ImgDescription: response.ImgDescription,
		SafeTitle:      response.SafeTitle,
	}, nil
}

// Получение ID последнего комикса с сайта XKCD
func (c Client) LastID(ctx context.Context) (int, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/info.0.json", c.url))
	if err != nil {
		c.log.Error(
			"failed to do get request for last comics",
			"url", c.url,
			"err", err,
		)
		return 0, core.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		c.log.Error(
			"failed to do get request for last comics",
			"url", c.url,
			"statusCode", resp.StatusCode,
		)
		return 0, core.ErrNotFound
	}
	//nolint:errcheck // close error because non-actionable here
	defer resp.Body.Close()

	var response responseLastID
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.log.Error(
			"failed to decode information from last comics",
			"err", err,
		)
		return 0, core.ErrNotFound
	}

	return response.Num, nil
}
