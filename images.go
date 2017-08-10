package dockerapi

import (
	"io"

	"github.com/fsouza/go-dockerclient"
)

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
	return c.Docker.PullImage(options, getAuthConfigurationFromDockerCfg())
}

// RemoveImage safely removes the image
func (c *Client) RemoveImage(image string) error {
	return c.Docker.RemoveImage(image)
}

// ImageExists checks if an image exists on the server
func (c *Client) ImageExists(image string) bool {
	_, err := c.Docker.InspectImage(image)
	return err == nil
}

func getAuthConfigurationFromDockerCfg() docker.AuthConfiguration {
	auth := docker.Authentication{}
	auths := docker.NewAuthConfigurationsFromDockerCfg()
	if len(auths) > 0 {
		for _, value := range auths {
			auth = value
			break // can't presume which docker config to use, so take the first one
		}
	}
	return auth
}
