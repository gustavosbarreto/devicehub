package client

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"regexp"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	DeviceUIDHeader = "X-Device-UID"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrNotFound         = errors.New("not found")
	ErrUnknown          = errors.New("unknown error")
)

// NewClient creates a new ShellHub HTTP client.
//
// Server address must contain the scheme, the host and the port. For instance: `https://cloud.shellhub.io:443/`.
func NewClient(address string, opts ...Opt) (Client, error) {
	uri, err := url.Parse(address)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not parse the address to the required format"), err)
	}

	client := new(client)
	client.http = resty.New()
	client.http.SetRetryCount(math.MaxInt32)
	client.http.SetRedirectPolicy(SameDomainRedirectPolicy())
	client.http.SetBaseURL(uri.String())
	client.http.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok {
			return true
		}

		log.WithFields(log.Fields{
			"status_code": r.StatusCode(),
			"url":         r.Request.URL,
		}).Warn("failed to achieve the server")

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	if client.logger != nil {
		client.http.SetLogger(&LeveledLogger{client.logger})
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
}

type client struct {
	scheme string
	host   string
	port   int
	http   *resty.Client
	logger *logrus.Logger
}

func (c *client) ListDevices() ([]models.Device, error) {
	list := []models.Device{}
	_, err := c.http.R().
		SetResult(&list).
		Get("/api/devices")
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	var device *models.Device
	resp, err := c.http.R().
		SetResult(&device).
		Get(fmt.Sprintf("/api/devices/%s", uid))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case 400:
		return nil, ErrNotFound
	case 200:
		return device, nil
	default:
		return nil, ErrUnknown
	}
}

// parseToWS gets a HTTP URI and change its values to meet the WebSocket format.
func parseToWS(uri string) string {
	return regexp.MustCompile(`^http`).ReplaceAllString(uri, "ws")
}
