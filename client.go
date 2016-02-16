package dockerapi

import "github.com/fsouza/go-dockerclient"

// Client is the docker client for this API
type Client struct {
	Docker *docker.Client
}

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
