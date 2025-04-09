package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	mpi "github.com/nginx/agent/v3/api/grpc/mpi/v1"
	"github.com/nginx/agent/v3/internal/bus"
	sloggin "github.com/samber/slog-gin"
)

var _ bus.Plugin = (*AgentAPIPlugin)(nil)

type AgentAPIPlugin struct {
	apiAddress  string
	server      *gin.Engine
	messagePipe bus.MessagePipeInterface
	mutex       sync.Mutex
	healths     []*mpi.InstanceHealth
	test        []string
}

func NewAgentAPI() *AgentAPIPlugin {
	return &AgentAPIPlugin{
		apiAddress: "0.0.0.0:9011",
		healths:    []*mpi.InstanceHealth{},
	}
}

func (a *AgentAPIPlugin) Init(ctx context.Context, messagePipe bus.MessagePipeInterface) error {
	slog.DebugContext(ctx, "Starting Agent API plugin")

	a.messagePipe = messagePipe
	go a.Start()
	return nil
}

func (a *AgentAPIPlugin) Close(ctx context.Context) error {
	slog.InfoContext(ctx, "Closing file plugin")
	return nil
}

func (a *AgentAPIPlugin) Info() *bus.Info {
	return &bus.Info{
		Name: "api",
	}
}

func (a *AgentAPIPlugin) Process(ctx context.Context, msg *bus.Message) {
	switch msg.Topic {
	case bus.InstanceHealthTopic:
		a.test = append(a.test, "hello")
		slog.InfoContext(ctx, "Received instance health event")
		a.handleInstanceHealthTopic(ctx, msg)
		slog.InfoContext(ctx, "Handled instance health event", "", a.healths)
	}
}

func (a *AgentAPIPlugin) handleInstanceHealthTopic(ctx context.Context, msg *bus.Message) {
	if instances, ok := msg.Data.([]*mpi.InstanceHealth); ok {
		if len(instances) > 0 {
			a.healths = instances
		}
	}
	slog.InfoContext(ctx, "Received health topic message", "health", a.healths)
}

func (a *AgentAPIPlugin) Subscriptions() []string {
	return []string{
		bus.InstanceHealthTopic,
	}
}

func (a *AgentAPIPlugin) Start() {
	a.server = gin.New()
	listener, err := net.Listen("tcp", a.apiAddress)
	slog.Info("Error starting API server", "error", err)
	a.server.Use(gin.Recovery())
	a.server.UseRawPath = true

	handler := slog.NewTextHandler(
		os.Stderr,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	)

	logger := slog.New(handler)

	a.server.Use(sloggin.NewWithConfig(logger, sloggin.Config{DefaultLevel: slog.LevelDebug}))

	a.addAgentHealthEndpoint()
	errListen := a.server.RunListener(listener)
	if errListen != nil {
		slog.Error("Failed to start Agent API http server", "error", err)
	}
}

func (a *AgentAPIPlugin) addAgentHealthEndpoint() {
	a.server.GET("/health", func(c *gin.Context) {
		slog.Info("Health endpoint added", "health", a.healths)
		if a.healths == nil {
			c.JSON(http.StatusNotFound, nil)
		} else {
			a.mutex.Lock()
			c.JSON(http.StatusOK, a.healths)
			a.mutex.Unlock()
		}
	})
}
