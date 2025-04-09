package api

import (
	"context"
	"github.com/gin-gonic/gin"
	mpi "github.com/nginx/agent/v3/api/grpc/mpi/v1"
	"github.com/nginx/agent/v3/internal/bus"
	sloggin "github.com/samber/slog-gin"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var _ bus.Plugin = (*AgentAPIPlugin)(nil)

type AgentAPIPlugin struct {
	apiAddress  string
	server      *gin.Engine
	messagePipe bus.MessagePipeInterface
	mutex       sync.Mutex
	test        []string
	health      []Health
}

type Health struct {
	Status      int       `json:"health"`
	Type        string    `json:"type"`
	InstanceID  string    `json:"instance_id"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

func NewAgentAPI() *AgentAPIPlugin {
	return &AgentAPIPlugin{
		apiAddress: "0.0.0.0:9095",
		health:     make([]Health, 0),
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
		slog.InfoContext(ctx, "Handled instance health event", "", a.health)
	}
}

func (a *AgentAPIPlugin) handleInstanceHealthTopic(ctx context.Context, msg *bus.Message) {
	if instances, ok := msg.Data.([]*mpi.InstanceHealth); ok {
		if len(instances) > 0 {
			a.convertHealth(instances)
		}
	}
	slog.InfoContext(ctx, "Received health topic message", "health", a.health)
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
		slog.Info("Health endpoint added", "health", a.health)
		if a.health == nil {
			c.JSON(http.StatusNotFound, nil)
		} else {

			a.mutex.Lock()
			for _, h := range a.health {
				h.Timestamp = time.Now()
			}
			c.JSON(http.StatusOK, a.health)
			a.mutex.Unlock()
		}
	})
}

func (a *AgentAPIPlugin) convertHealth(instanceHealth []*mpi.InstanceHealth) {
	a.health = make([]Health, 0)
	for _, h := range instanceHealth {
		status := 0

		switch h.GetInstanceHealthStatus() {
		case mpi.InstanceHealth_INSTANCE_HEALTH_STATUS_HEALTHY:
			status = 1
		case mpi.InstanceHealth_INSTANCE_HEALTH_STATUS_DEGRADED:
			status = 3
		case mpi.InstanceHealth_INSTANCE_HEALTH_STATUS_UNHEALTHY:
			status = 2
		default:
			status = 0
		}

		newHealth := Health{
			Status:      status,
			Type:        h.GetInstanceType().String(),
			Timestamp:   time.Now(),
			InstanceID:  h.GetInstanceId(),
			Description: h.GetDescription(),
		}
		a.health = append(a.health, newHealth)
	}
}
