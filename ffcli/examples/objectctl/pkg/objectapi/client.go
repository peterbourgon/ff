package objectapi

import (
	"context"
	"errors"
	"time"
)

// Object is meant to be a domain object for a theoretical object store.
type Object struct {
	Key    string
	Value  string
	Access time.Time
}

// Client is meant to model an SDK client for a theoretical object store API.
// Because we're only using it for demo purposes, it embeds a mock server with
// fixed data.
type Client struct {
	token  string
	server *mockServer
}

// NewClient is meant to model a constructor for the SDK client.
func NewClient(token string) (*Client, error) {
	return &Client{
		token:  token,
		server: newMockServer(),
	}, nil
}

// Create is some bit of functionality.
func (c *Client) Create(ctx context.Context, key, value string, overwrite bool) error {
	return c.server.create(c.token, key, value, overwrite)
}

// Delete is some bit of functionality.
func (c *Client) Delete(ctx context.Context, key string, force bool) (existed bool, err error) {
	return c.server.delete(c.token, key, force)
}

// List is some bit of functionality.
func (c *Client) List(ctx context.Context) ([]Object, error) {
	return c.server.list(c.token)
}

//
//
//

type mockServer struct {
	token   string
	objects map[string]Object
}

func newMockServer() *mockServer {
	return &mockServer{
		token:   "SECRET",
		objects: defaultObjects,
	}
}

func (s *mockServer) create(token, key, value string, overwrite bool) error {
	if token != s.token {
		return errors.New("not authorized")
	}

	if _, ok := s.objects[key]; ok && !overwrite {
		return errors.New("object already exists")
	}

	s.objects[key] = Object{
		Key:    key,
		Value:  value,
		Access: time.Now(),
	}

	return nil
}

func (s *mockServer) delete(token, key string, force bool) (existed bool, err error) {
	if token != s.token {
		return false, errors.New("not authorized")
	}

	_, ok := s.objects[key]
	delete(s.objects, key)
	return ok, nil
}

func (s *mockServer) list(token string) ([]Object, error) {
	if token != s.token {
		return nil, errors.New("not authorized")
	}

	result := make([]Object, 0, len(s.objects))
	for _, obj := range s.objects {
		result = append(result, obj)
	}

	return result, nil
}

var defaultObjects = map[string]Object{
	"apple": {
		Key:    "apple",
		Value:  "The fruit of any of certain other species of tree of the same genus.",
		Access: mustParseTime(time.RFC3339, "2019-03-15T15:01:00Z"),
	},
	"beach": {
		Key:    "beach",
		Value:  "The shore of a body of water, especially when sandy or pebbly.",
		Access: mustParseTime(time.RFC3339, "2019-04-20T12:21:30Z"),
	},
	"carillon": {
		Key:    "carillon",
		Value:  "A stationary set of chromatically tuned bells in a tower.",
		Access: mustParseTime(time.RFC3339, "2019-07-04T23:59:59Z"),
	},
}

func mustParseTime(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}
