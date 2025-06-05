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
	"testing"

	"github.com/kubetail-org/kubetail/modules/shared/config"
	k8shelpersmock "github.com/kubetail-org/kubetail/modules/shared/k8shelpers/mock"
	"github.com/stretchr/testify/assert"
)

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
