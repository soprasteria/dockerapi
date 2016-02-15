package dockerapi

import (
	"io"

	"github.com/fsouza/go-dockerclient"
)

// NewClient creates a new Docker client
func NewClient(endpoint string) (*Client, error) {
	c, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

// NewTLSClient create a client for a TLS secured Docker engine
func NewTLSClient(host, certPEM, keyPEM, caPEM string) (*Client, error) {
	c, err := docker.NewTLSClient(host, certPEM, keyPEM, caPEM)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

// PullImage pulls an Docker image
func (c *Client) PullImage(image string) error {
	return c.PullImageAsync(image, nil)
}

// PullImageAsync pull the given image and progress can be followed asynchronously, by providing a writer
func (c *Client) PullImageAsync(image string, progressDetail io.Writer) error {
	options := docker.PullImageOptions{
		Repository:   image,
		OutputStream: progressDetail,
	}
	auth := docker.AuthConfiguration{}
	return c.Docker.PullImage(options, auth)
}
