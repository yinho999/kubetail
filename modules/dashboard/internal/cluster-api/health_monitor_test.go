// Copyright 2024-2025 Andres Morey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clusterapi

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kubetail-org/kubetail/modules/shared/config"
	k8shelpersmock "github.com/kubetail-org/kubetail/modules/shared/k8shelpers/mock"
	"github.com/stretchr/testify/assert"
)

type MockHealthMonitorWorker struct {
	showdownTime time.Duration
	shutdown     atomic.Bool
}

func newMockHealthMonitorWorker(showdownTime time.Duration) *MockHealthMonitorWorker {
	return &MockHealthMonitorWorker{
		shutdown:     atomic.Bool{},
		showdownTime: showdownTime,
	}
}

func (m *MockHealthMonitorWorker) Start(ctx context.Context) error {
	return nil
}

func (m *MockHealthMonitorWorker) Shutdown() {
	m.shutdown.Store(true)
}

func (m *MockHealthMonitorWorker) GetHealthStatus() HealthStatus {
	return HealthStatusSuccess
}

func (m *MockHealthMonitorWorker) WatchHealthStatus(ctx context.Context) (<-chan HealthStatus, error) {
	return nil, nil
}

func (m *MockHealthMonitorWorker) ReadyWait(ctx context.Context) error {
	return nil
}

func TestNewHealthMonitor(t *testing.T) {
	tests := []struct {
		name        string
		environment config.Environment
		endpoint    string
		want        interface{}
	}{
		{
			name:        "Desktop environment",
			environment: config.EnvironmentDesktop,
			endpoint:    "",
			want:        &DesktopHealthMonitor{},
		},
		{
			name:        "Cluster environment",
			environment: config.EnvironmentCluster,
			endpoint:    "",
			want:        &InClusterHealthMonitor{},
		},
		{
			name:        "Cluster environment with endpoint",
			environment: config.EnvironmentCluster,
			endpoint:    "kubetail-cluster-api.kubetail-system.svc.cluster.local:50051",
			want:        &InClusterHealthMonitor{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create connection manager
			cm := &k8shelpersmock.MockConnectionManager{}

			// create config
			cfg := config.DefaultConfig()
			cfg.Dashboard.Environment = tt.environment
			if tt.endpoint != "" {
				cfg.Dashboard.ClusterAPIEndpoint = tt.endpoint
			}

			// create health monitor
			hm := NewHealthMonitor(cfg, cm)

			// assert health monitor is not nil
			assert.NotNil(t, hm)

			// assert health monitor is of the expected type
			assert.IsType(t, tt.want, hm)

		})
	}

}

func TestNewHealthMonitor_InvalidEnvironment(t *testing.T) {
	// create connection manager
	cm := &k8shelpersmock.MockConnectionManager{}

	// create config
	cfg := config.DefaultConfig()
	cfg.Dashboard.Environment = "invalid"

	// Assert panic
	assert.Panics(t, func() {
		NewHealthMonitor(cfg, cm)
	}, "NewHealthMonitor should panic if environment is invalid")
}

// Shutdown
func TestDesktopHealthMonitor_Shutdown(t *testing.T) {

	tests := []struct {
		name          string
		numberWorkers int
		showdownTime  time.Duration
	}{
		{
			name:          "Shutdown with 0 workers",
			numberWorkers: 0,
			showdownTime:  1 * time.Second,
		},
		{
			name:          "Shutdown with 1 worker",
			numberWorkers: 2,
			showdownTime:  1 * time.Second,
		},
		{
			name:          "Shutdown with multiple workers",
			numberWorkers: 2,
			showdownTime:  1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create connection manager
			cm := &k8shelpersmock.MockConnectionManager{}

			// create health monitor
			hm := NewDesktopHealthMonitor(cm)

			var workers []*MockHealthMonitorWorker

			// create mock health monitor worker
			for i := 0; i < tt.numberWorkers; i++ {
				worker := newMockHealthMonitorWorker(tt.showdownTime)
				workers = append(workers, worker)
				hm.workerCache.Store(fmt.Sprintf("worker-%d", i), worker)
			}

			// shutdown health monitor
			hm.Shutdown()

			// assert worker is shutdown
			for _, worker := range workers {
				assert.True(t, worker.shutdown.Load())
			}
		})
	}
}

func TestInClusterHealthMonitor_Shutdown(t *testing.T) {

	tests := []struct {
		name         string
		endpoint     string
		hasWorker    bool
		showdownTime time.Duration
		want         interface{}
	}{
		{
			name:         "Shutdown with endpoint and 0 workers",
			endpoint:     "kubetail-cluster-api.kubetail-system.svc.cluster.local:50051",
			hasWorker:    false,
			showdownTime: 1 * time.Second,
			want:         &InClusterHealthMonitor{},
		},
		{
			name:         "Shutdown with endpoint and 1 worker",
			endpoint:     "kubetail-cluster-api.kubetail-system.svc.cluster.local:50051",
			hasWorker:    true,
			showdownTime: 1 * time.Second,
			want:         &InClusterHealthMonitor{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create connection manager
			cm := &k8shelpersmock.MockConnectionManager{}

			// create health monitor
			hm := NewInClusterHealthMonitor(cm, tt.endpoint)
			var worker *MockHealthMonitorWorker

			// create mock health monitor worker
			if tt.hasWorker {
				worker = newMockHealthMonitorWorker(tt.showdownTime)
				hm.worker = worker
			}

			// assert health monitor is not nil
			assert.NotNil(t, hm)

			// assert health monitor is of the expected type
			assert.IsType(t, tt.want, hm)

			// shutdown health monitor
			hm.Shutdown()

			// assert worker is shutdown
			if tt.hasWorker {
				assert.True(t, worker.shutdown.Load())
			}

		})
	}
}
