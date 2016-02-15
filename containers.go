package dockerapi

import (
	"errors"
	"fmt"
	"log"

	"github.com/fsouza/go-dockerclient"
)

// ContainerOptions defines options for container initialisation
type ContainerOptions struct {
	Image        string        // Name of the image in the registry (ex : redis:latest)
	Name         string        // Name of the container
	PortBindings []PortBinding // List of ports to bind
}

// NewContainer initializes a new container, ready to be created
func (c *Client) NewContainer(o ContainerOptions) (*Container, error) {
	if c == nil {
		return nil, errors.New("Docker client is not initialized")
	}
	if o.Image == "" {
		return nil, errors.New("Image is required")
	}
	if o.Name == "" {
		return nil, errors.New("Name is required")
	}

	return &Container{
		Name:         o.Name,
		Image:        o.Image,
		PortBindings: o.PortBindings,
		Client:       c,
	}, nil
}

// Create creates the container
func (c *Container) Create() error {

	portBindings := map[docker.Port][]docker.PortBinding{}
	for _, binding := range c.PortBindings {
		port := docker.Port(binding.ContainerPort + "/tcp")
		portBindings[port] = []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: binding.HostPort}}
	}

	cont, err := c.Client.Docker.CreateContainer(docker.CreateContainerOptions{
		Name:       c.Name,
		Config:     &docker.Config{Image: c.Image},
		HostConfig: &docker.HostConfig{PortBindings: portBindings},
	})
	c.Container = cont

	return err
}

// Start starts the container
func (c *Container) Start() error {
	err := c.Client.Docker.StartContainer(c.Container.ID, c.Container.HostConfig)
	if err != nil {
		return fmt.Errorf("Can't start container %v because %v", c.Container.ID, err.Error())
	}
	return nil
}

// Run runs the container, aka pull image, create, start
func (c *Container) Run() error {

	log.Printf("Pulling %+v image\n", c.Image)
	err := c.Client.PullImage(c.Image)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Unable to donwload %v image", c.Image)
	}

	log.Printf("Creating container %+v\n", c.Name)
	err = c.Create()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Can't create container %+v", c.Name)
	}

	log.Printf("Starting container %+v\n", c.Name)
	err = c.Start()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Can't start %+v", c.Name)
	}

	log.Printf("Container %v is started with id %v", c.Container.Name, c.Container.ID)

	return nil
}

// Stop stops a container
func (c *Container) Stop() error {
	id := c.Container.ID
	if id == "" {
		return fmt.Errorf("Container %+v does not exist", c)
	}
	err := c.Client.Docker.StopContainer(id, 30)
	if err != nil {
		return fmt.Errorf("Can't stop container of id:%v (%v)", id, err.Error())
	}
	return nil
}

// Remove removes a container,
// Volumes is a flag indicating whether Docker should remove the volumes associated to the container.
func (c *Container) Remove(volumes bool) error {
	id := c.Container.ID
	if id == "" {
		return fmt.Errorf("Container %+v does not exist", c)
	}
	options := docker.RemoveContainerOptions{
		ID:            c.Container.ID,
		Force:         false,
		RemoveVolumes: volumes,
	}
	err := c.Client.Docker.RemoveContainer(options)
	if err != nil {
		options.Force = true
		err = c.Client.Docker.RemoveContainer(options)
		if err != nil {
			return fmt.Errorf("Can't remove container of id %v", c.Container.ID)
		}
	}
	return nil
}

// StopAndRemove stop and remove the container and possibly its volumes
// Returns error if something bad happened
func (c *Container) StopAndRemove(volumes bool) error {
	err := c.Stop()
	if err != nil {
		return err
	}
	err = c.Remove(volumes)
	return err
}

// RunAll runs all containers from the pool
// Returns error if something bad happened but no error exits
func (pool PoolContainer) RunAll() (err error) {
	sem := make(chan error, len(pool))
	// Concurrent Run
	for _, v := range pool {
		go func(v *Container) {
			sem <- v.Run()
		}(v)
	}
	// Waiting for return
	for i := 0; i < len(pool); i++ {
		err = <-sem
		if err != nil {
			log.Println(err)
		}
	}
	return
}

// RemoveAll stops and remove all containers from the pool
// Returns error if something bad happened but no error exits
func (pool PoolContainer) RemoveAll(volumes bool) (err error) {

	// Concurrent Remove
	sem := make(chan error, len(pool))
	for _, v := range pool {
		go func(v *Container) {
			sem <- v.Remove(volumes)
		}(v)
	}
	// Waiting for return
	for i := 0; i < len(pool); i++ {
		err = <-sem
		if err != nil {
			log.Println(err)
		}
	}
	return err
}
