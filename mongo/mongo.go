package mongo

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	tc "github.com/tordf/testcontainers"
)

var mongoMux = new(sync.Mutex)

// ContainerOptions ...
type ContainerOptions struct {
	tc.ContainerOptions
	User     string
	Password string
}

// DBConfig ...
type DBConfig struct {
	tc.ContainerConfig
	Host     string
	Port     uint
	User     string
	Password string
}

// ConnectionURI ...
func (c DBConfig) ConnectionURI() string {
	var databaseAuth string
	if c.User != "" && c.Password != "" {
		databaseAuth = fmt.Sprintf("%s:%s@", c.User, c.Password)
	}
	databaseHost := fmt.Sprintf("%s:%d", c.Host, c.Port)
	return fmt.Sprintf("mongodb://%s%s/?connect=direct", databaseAuth, databaseHost)
}

const defaultMongoDBPort = 27017

// StartMongoContainer ...
func StartMongoContainer(ctx context.Context, options ContainerOptions) (mongoC testcontainers.Container, Config DBConfig, err error) {
	mongoMux.Lock()
	defer mongoMux.Unlock()
	mongoPort, _ := nat.NewPort("", strconv.Itoa(defaultMongoDBPort))

	env := make(map[string]string)
	if options.User != "" && options.Password != "" {
		env["MONGO_INITDB_ROOT_USERNAME"] = options.User
		env["MONGO_INITDB_ROOT_PASSWORD"] = options.Password
	}

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	req := testcontainers.ContainerRequest{
		Image:        "mongo:4.4.3",
		Env:          env,
		ExposedPorts: []string{string(mongoPort)},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(timeout),
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	tc.ClientMux.Lock()
	mongoC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.ClientMux.Unlock()
	if err != nil {
		err = fmt.Errorf("Failed to start mongo container: %v", err)
		return
	}

	host, err := mongoC.Host(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to get mongo container host: %v", err)
		return
	}

	port, err := mongoC.MappedPort(ctx, mongoPort)
	if err != nil {
		err = fmt.Errorf("Failed to get exposed mongo container port: %v", err)
		return
	}

	Config = DBConfig{
		Host:     host,
		Port:     uint(port.Int()),
		User:     options.User,
		Password: options.Password,
	}

	if options.CollectLogs {
		Config.ContainerConfig.Log = &tc.LogCollector{
			MessageChan: make(chan string),
			Mux:         sync.Mutex{},
		}
		go tc.EnableLogger(mongoC, Config.ContainerConfig.Log)
	}
	return
}
