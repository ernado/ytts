package speechkit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/go-faster/errors"
	"github.com/google/go-querystring/query"
)

type HTTP interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	http     HTTP
	folderID string
	token    string
}

type options struct {
	http     HTTP
	folderID string
}

type Option interface {
	apply(o *options)
}

type optionFunc func(o *options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithHTTP sets the HTTP client to use.
func WithHTTP(h HTTP) Option {
	return optionFunc(func(o *options) {
		o.http = h
	})
}

// WithFolderID sets the folder to use.
func WithFolderID(id string) Option {
	return optionFunc(func(o *options) {
		o.folderID = id
	})
}

// New initializes a new speechkit client.
func New(token string, opts ...Option) *Client {
	o := options{
		http: http.DefaultClient,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return &Client{
		http:     o.http,
		folderID: o.folderID,
		token:    token,
	}
}

type Options struct {
	Text     string  `url:"text"`
	Language string  `url:"lang"`    // ru-RU, en-US, ...
	Speed    float64 `url:"speed"`   // 0.1 - 3.0
	Voice    string  `url:"voice"`   // omazh (f), zahar (m), jane (f)
	Emotion  string  `url:"emotion"` // good, neutral, evil
}

type Error struct {
	Code    string `json:"error_code"`
	Message string `json:"error_message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// https://cloud.yandex.ru/services/speechkit#demo
// https://cloud.yandex.ru/docs/speechkit/tts/request
const apiURL = "https://tts.api.cloud.yandex.net/speech/v1/tts:synthesize"

// Synthesize synthesizes speech from the given text.
func (c *Client) Synthesize(ctx context.Context, opt Options) (io.ReadCloser, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		panic(err)
	}
	v, err := query.Values(opt)
	if err != nil {
		panic(err)
	}
	if c.folderID != "" {
		v.Set("folderId", c.folderID)
	}
	u.RawQuery = v.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), http.NoBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("YANDEX_TOKEN"))

	res, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}
	if res.StatusCode != http.StatusOK {
		defer func() { _ = res.Body.Close() }()
		d := json.NewDecoder(res.Body)
		var e Error
		if err := d.Decode(&e); err != nil || e.Code == "" {
			return nil, errors.Errorf("unexpected code %d", res.StatusCode)
		}
		return nil, &e
	}

	return res.Body, nil
}
