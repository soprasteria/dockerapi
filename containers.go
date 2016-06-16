package dockerapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/soprasteria/dockerapi/utils"

	"github.com/fsouza/go-dockerclient"
)

// Container is a docker container
type Container struct {
	Container *docker.Container // fsouza docker client. To use if this wrapper is not able to do what you want
	Client    *Client           // wrapper client used to create the container. Will be used for any other Docker action
}

// PortBinding binds the port from host and container from host
type PortBinding struct {
	ContainerPort string // Port inside the container
	HostPort      string // port outside the container (on the host)
	Protocol      string // tcp/udp
}

// Parameters list all docker parameters (for example, to limit the docker container : memory, cpu etc.)
type Parameters struct {
	Memory     int64
	MemorySwap int64
	CPUShares  int64
	CPUSet     string
}

// ContainerOptions defines options for container initialisation
type ContainerOptions struct {
	Image        string        // Name of the image in the registry (ex : redis:latest)
	Name         string        // Name of the container
	PortBindings []PortBinding // List of ports to bind
	Cmd          []string      // command to launch when starting the container
	Binds        []string      // Volume bindings. Format :  externalpath:internalpath:r(w|o)
	Links        []string      // Links to use inside the container. Format : externalname:internalname
	Env          []string      // Environment variables to set for the container. Format : key=value
	Hostname     string        // Hostname of the docker container
	Parameters   Parameters    // Parameters list all docker parameters
}

// NewContainer initializes a new container, ready to be created
func (c *Client) NewContainer(o ContainerOptions) (*Container, error) {
	if o.Image == "" {
		return nil, errors.New("Image is required")
	}
	if o.Name == "" {
		return nil, errors.New("Name is required")
	}

	// Handle port bindings and default behaviour
	portBindings := map[docker.Port][]docker.PortBinding{}
	for _, binding := range o.PortBindings {
		if binding.Protocol == "" || binding.Protocol != "udp" {
			// TCP port by default
			binding.Protocol = "tcp"
		}
		port := docker.Port(binding.ContainerPort + "/" + binding.Protocol)
		portBindings[port] = []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: binding.HostPort}}
	}

	// Handle volume bindings and default behaviour
	volumeBindings := []string{}
	for _, binding := range o.Binds {
		volume := strings.Split(binding, ":")
		if len(volume) == 2 {
			// external:internal -> external:internal:rw
			binding = binding + ":rw"
		}

		volumeBindings = append(volumeBindings, binding)
	}

	container := &docker.Container{
		Name: o.Name,
		Config: &docker.Config{
			Image:    o.Image,
			Cmd:      o.Cmd,
			Env:      o.Env,
			Hostname: o.Hostname,
		},
		HostConfig: &docker.HostConfig{
			PortBindings: portBindings,
			Binds:        volumeBindings,
			Links:        o.Links,
			Memory:       o.Parameters.Memory,
			MemorySwap:   o.Parameters.MemorySwap,
			CPUShares:    o.Parameters.CPUShares,
			CPUSet:       o.Parameters.CPUSet,
		},
	}

	return &Container{
		Container: container,
		Client:    c,
	}, nil
}

// InspectContainer inspects the container on server from an id
func (c *Client) InspectContainer(id string) (*Container, error) {
	cont, err := c.Docker.InspectContainer(id)
	if err != nil {
		return nil, err
	}
	return &Container{
		Container: cont,
		Client:    c,
	}, nil
}

// ListRunningContainers list all running containers on docker engine
func (c *Client) ListRunningContainers() (res LightContainers, err error) {

	options := docker.ListContainersOptions{}
	containers, err := c.Docker.ListContainers(options)
	if err != nil {
		return
	}

	for _, v := range containers {
		res = append(res, LightContainer{v})
	}
	return
}

// ListContainers list all running and non-running containers on docker engine
func (c *Client) ListContainers() (res LightContainers, err error) {
	options := docker.ListContainersOptions{All: true}
	containers, err := c.Docker.ListContainers(options)
	if err != nil {
		return
	}

	for _, v := range containers {
		res = append(res, LightContainer{v})
	}
	return
}

// LightContainer is a simple docker container returned by ListContainers
type LightContainer struct {
	Container docker.APIContainers
}

// ID returns the id of the light container
func (c LightContainer) ID() string {
	return c.Container.ID
}

// ShortID returns a short representation of an id of the light container
func (c LightContainer) ShortID() string {
	return utils.SubString(c.ID(), 12)
}

// Image returns the image of the light container
func (c LightContainer) Image() string {
	return c.Container.Image
}

// LightContainers is a slice of LightContainer
type LightContainers []LightContainer

//GetIDs returns a slice of ids from a slice of light containers
func (containers LightContainers) GetIDs() []string {
	result := []string{}
	for _, c := range containers {
		result = append(result, c.ID())
	}
	return result
}

// Name returns the name of the container
func (c *Container) Name() (name string) {
	if c.Container != nil {
		name = c.Container.Name
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
	}
	return
}

// ID returns the id of the container
func (c *Container) ID() string {
	if c.Container != nil {
		return c.Container.ID
	}
	return ""
}

// ShortID returns a short representation of an id of the container
func (c *Container) ShortID() string {
	return utils.SubString(c.ID(), 12)
}

// Image returns the image of the container
func (c *Container) Image() (image string) {
	if c.Container != nil && c.Container.Config != nil {
		image = c.Container.Config.Image
	}
	return
}

// IsRunning checks that container is running
func (c *Container) IsRunning() bool {
	if c.Container != nil {
		return c.Container.State.Running
	}
	return false
}

// GetEnvs returns the list of environment variables inside the container
func (c *Container) GetEnvs() []string {
	if c.Container != nil {
		return c.Container.Config.Env
	}

	return []string{}
}

// Rename renames a container's name to another
func (c *Container) Rename(newName string) error {
	if newName == "" {
		return errors.New("New name is empty")
	}
	if newName[0] == '/' {
		newName = newName[1:]
	}

	options := docker.RenameContainerOptions{
		ID:   c.ID(),
		Name: newName,
	}
	err := c.Client.Docker.RenameContainer(options)

	if err != nil {
		return fmt.Errorf("Can't rename %v to %v because %v", c.Name(), newName, err.Error())
	}

	return c.Refresh()
}

// Clone clones an existing container configuration
// Uses golang gob to serialize and deserialise the object, in order to get a deep copy of the object
func (c *Container) Clone() (*Container, error) {

	// Create an encoder and send a value.
	marsh, err := json.Marshal(*c.Container)
	if err != nil {
		return nil, fmt.Errorf("Can't encode container %v : %v", c.ShortID(), err)
	}

	// Create a decoder and receive a value.
	var clone docker.Container
	err = json.Unmarshal(marsh, &clone)
	if err != nil {
		return nil, fmt.Errorf("Can't decode container %v : %v", c.ShortID(), err)
	}
	clone.ID = ""

	return &Container{
		Container: &clone,
		Client:    c.Client,
	}, nil

}

// Refresh refresh container from the server
func (c *Container) Refresh() error {
	cont, err := c.Client.InspectContainer(c.Container.ID)
	if err != nil {
		return err
	}
	c.Container = cont.Container
	return nil
}

// Create creates the container
func (c *Container) Create() error {
	cont, err := c.Client.Docker.CreateContainer(docker.CreateContainerOptions{
		Name:       c.Container.Name,
		Config:     c.Container.Config,
		HostConfig: c.Container.HostConfig,
	})
	if err != nil {
		return err
	}
	c.Container = cont
	return err
}

// Start starts the container
func (c *Container) Start() error {
	err := c.Client.Docker.StartContainer(c.Container.ID, c.Container.HostConfig)
	if err != nil {
		return fmt.Errorf("Can't start container %v because %v", c.ShortID(), err.Error())
	}
	c.Refresh()
	return nil
}

// Run runs the container, aka pull image, create, start
func (c *Container) Run() error {

	log.Printf("Pulling %+v image\n", c.Image())
	err := c.Client.PullImage(c.Image())
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Unable to donwload %v image", c.Image())
	}

	log.Printf("Creating container %+v\n", c.Name())
	err = c.Create()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Can't create container %+v", c.Name())
	}

	log.Printf("Starting container %+v\n", c.Name())
	err = c.Start()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Can't start %+v", c.Name())
	}

	log.Printf("Container %v is started with id %v", c.Name(), c.ShortID())

	return nil
}

// Stop stops a container
func (c *Container) Stop() error {
	err := c.Client.Docker.StopContainer(c.Container.ID, 30)
	if err != nil {
		return fmt.Errorf("Can't stop container of id:%v (%v)", c.ShortID(), err.Error())
	}
	c.Refresh()
	return nil
}

// Remove removes a container,
// Volumes is a flag indicating whether Docker should remove the volumes associated to the container.
func (c *Container) Remove(volumes bool) error {

	// Remove the container gracefull, then by force
	// First to

	superRemove := func(id string, volumes bool) error {

		var err error

		if id != "" {
			options := docker.RemoveContainerOptions{
				ID:            id,
				Force:         false,
				RemoveVolumes: volumes,
			}
			// Graceful removal
			err = c.Client.Docker.RemoveContainer(options)
			if err == nil {
				return nil
			}
			// Forced removal
			options.Force = true
			err = c.Client.Docker.RemoveContainer(options)
			if err == nil {
				return nil
			}
		} else {
			err = errors.New("ID is empty")
		}
		return fmt.Errorf("Can't remove container with id %v -> %v)", id, err)
	}

	err := superRemove(c.ID(), volumes)
	if err == nil {
		return nil
	}
	err = superRemove(c.Name(), volumes)
	if err == nil {
		return nil
	}

	return fmt.Errorf("Can't remove container %v (%v). Error : %v", c.Name(), c.ShortID(), err)
}

// StopAndRemove stop and remove the container and possibly its volumes
// Returns error if something bad happened
func (c *Container) StopAndRemove(volumes bool) error {
	err := c.Stop()
	if err != nil {
		return err
	}
	return c.Remove(volumes)
}

// ExecSh executes a command in sh shell on a container
func (c *Container) ExecSh(cmd []string) (logs []string, err error) {
	bash := []string{"/bin/sh", "-c"}
	return c.Exec(append(bash, cmd...))
}

// Exec executes a command on a container
func (c *Container) Exec(cmd []string) (logs []string, err error) {
	if c.ID() == "" {
		return logs, fmt.Errorf("Container %+v does not exist", c)
	}

	r, w := io.Pipe()
	success := make(chan struct{})
	createOptions := docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          cmd,
		Container:    c.ID(),
	}
	execOptions := docker.StartExecOptions{
		Detach:       false,
		Tty:          false,
		OutputStream: w,
		ErrorStream:  w,
		RawTerminal:  false,
		Success:      success,
	}
	exec, err := c.Client.Docker.CreateExec(createOptions)
	if err != nil {
		return logs, err
	}

	go func() {
		defer r.Close()
		if errr := c.Client.Docker.StartExec(exec.ID, execOptions); errr != nil {
			log.Fatal(errr)
			err = errr
		}
	}()
	<-success
	close(success)
	if err != nil {
		return logs, err
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}

	execInspect, err := c.Client.Docker.InspectExec(exec.ID)
	if err != nil {
		return logs, err
	}
	if execInspect.ExitCode != 0 {
		return logs, fmt.Errorf("Command %q failed : %v ", strings.Join(cmd, " "), execInspect.ExitCode)
	}

	return logs, nil
}

// LogsOptions is used to get logs from container
type LogsOptions struct {
	OutputStream io.Writer
	ErrorStream  io.Writer
	Stdout       bool
	Stderr       bool
	Tail         string
}

// Logs get the logs from the container
func (c *Container) Logs(opts LogsOptions) error {
	err := c.Client.Docker.Logs(docker.LogsOptions{
		Container:    c.ID(),
		OutputStream: opts.OutputStream,
		ErrorStream:  opts.ErrorStream,
		Stderr:       opts.Stderr,
		Stdout:       opts.Stdout,
		Tail:         opts.Tail,
		Follow:       true,
		Timestamps:   true,
	})
	if err != nil {
		return fmt.Errorf("Can't get logs from container %v because : %v", c.ShortID(), err.Error())
	}
	return nil
}

// PoolContainer is a pool of container. Can do mass operations on this
type PoolContainer []*Container

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
