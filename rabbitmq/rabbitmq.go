package rabbitmq

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	tc "github.com/tordf/testcontainers"
)

// ContainerOptions ...
type ContainerOptions struct {
	tc.ContainerOptions
}

// Config ...
type Config struct {
	tc.ContainerConfig
	Host string
	Port int64
}

const (
	defaultRabbitmqPort = 5672
)

// StartRabbitmqContainer ...
func StartRabbitmqContainer(ctx context.Context, options ContainerOptions) (rabbitmqC testcontainers.Container, rabbitmqConfig Config, err error) {
	rabbitmqPort, _ := nat.NewPort("", strconv.Itoa(defaultRabbitmqPort))

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.8.11-management",
		ExposedPorts: []string{string(rabbitmqPort)},
		WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(timeout),
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	tc.ClientMux.Lock()
	rabbitmqC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.ClientMux.Unlock()
	if err != nil {
		err = fmt.Errorf("Failed to start rabbitmq container: %v", err)
		return
	}

	host, err := rabbitmqC.Host(ctx)
	if err != nil {
		err = fmt.Errorf("Failed to get rabbitmq container host: %v", err)
		return
	}

	port, err := rabbitmqC.MappedPort(ctx, rabbitmqPort)
	if err != nil {
		err = fmt.Errorf("Failed to get exposed rabbitmq container port: %v", err)
		return
	}

	rabbitmqConfig = Config{
		Host: host,
		Port: int64(port.Int()),
	}

	if options.CollectLogs {
		rabbitmqConfig.ContainerConfig.Log = new(tc.LogCollector)
		go tc.EnableLogger(rabbitmqC, rabbitmqConfig.ContainerConfig.Log)
	}
	return
}
