// Copyright (c) F5, Inc.
//
// This source code is licensed under the Apache License, Version 2.0 license found in the
// LICENSE file in the root directory of this source tree.

package health

import (
	"context"
	"fmt"
	"sync"

	mpi "github.com/nginx/agent/v3/api/grpc/mpi/v1"
)

type AgentHealthWatcher struct {
	connected bool
	mutex     sync.Mutex
}

var _ healthWatcherOperator = (*AgentHealthWatcher)(nil)

func NewAgentHealthWatcher() *AgentHealthWatcher {
	return &AgentHealthWatcher{
		connected: false,
	}
}

func (ahw *AgentHealthWatcher) Health(_ context.Context, instance *mpi.Instance) (*mpi.InstanceHealth, error) {
	health := &mpi.InstanceHealth{
		InstanceId:           instance.GetInstanceMeta().GetInstanceId(),
		InstanceHealthStatus: mpi.InstanceHealth_INSTANCE_HEALTH_STATUS_HEALTHY,
		InstanceType:         instance.GetInstanceMeta().GetInstanceType(),
	}
	ahw.mutex.Lock()
	if !ahw.connected {
		health.InstanceHealthStatus = mpi.InstanceHealth_INSTANCE_HEALTH_STATUS_DEGRADED
		health.Description = fmt.Sprintf("Agent is not currentlty connected to a Management Plane")
	}
	ahw.mutex.Unlock()
	return health, nil
}

func (ahw *AgentHealthWatcher) SetConnected(connected bool) {
	ahw.mutex.Lock()
	ahw.connected = connected
	ahw.mutex.Unlock()
}
