package dockerapi

import "github.com/fsouza/go-dockerclient"

// Client is the docker client for this API
type Client struct {
	Docker *docker.Client
}

// PortBinding binds the port from host and container from host
type PortBinding struct {
	ContainerPort string // Port inside the container
	HostPort      string // port outside the container (on the host)
}

// Container is a docker container
type Container struct {
	Container    *docker.Container // fsouza docker client. To use if this wrapper is not able to do what you want
	Client       *Client           // wrapper client used to create the container. Will be used for any other Docker action
	Image        string            // Image like "redis:latest"
	Name         string            // Name when created
	PortBindings []PortBinding     // Binding 'hostport:containerport'
}

// PoolContainer is a pool of container. Can do mass operations on this
type PoolContainer []*Container
