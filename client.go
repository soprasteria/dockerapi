package dockerapi

import "github.com/fsouza/go-dockerclient"

// Client is the docker client for this API
type Client struct {
	Docker *docker.Client
}

type TLSClientFromBytesParameters struct {
	Host                                 string
	CertPEMBlock, KeyPEMBlock, CaPEMCert []byte
	InsecureSkipVerify                   bool
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
// The key and certificates are passed by filename
func NewTLSClient(host, certPEM, keyPEM, caPEM string) (*Client, error) {
	c, err := docker.NewTLSClient(host, certPEM, keyPEM, caPEM)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

// NewTLSClientFromBytes create a client for a TLS secured Docker engine
// The key and certificates are passed inline
func NewTLSClientFromBytes(params TLSClientFromBytesParameters) (*Client, error) {
	c, err := docker.NewTLSClientFromBytes(params.Host, params.CertPEMBlock, params.KeyPEMBlock, params.CaPEMCert)
	if err != nil {
		return nil, err
	}
	c.TLSConfig.InsecureSkipVerify = params.InsecureSkipVerify
	return &Client{c}, nil
}
