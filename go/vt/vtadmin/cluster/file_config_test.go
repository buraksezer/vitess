package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestFileConfigUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name   string
		yaml   string
		config FileConfig
		err    error
	}{
		{
			name: "simple",
			yaml: `defaults:
    discovery: consul
    discovery-consul-vtgate-datacenter-tmpl: "dev-{{ .Cluster }}"
    discovery-consul-vtgate-service-name: vtgate-svc
    discovery-consul-vtgate-pool-tag: type
    discovery-consul-vtgate-cell-tag: zone
    discovery-consul-vtgate-addr-tmpl: "{{ .Name }}.example.com:15999"

clusters:
    c1:
        name: testcluster1
        discovery-consul-vtgate-datacenter-tmpl: "dev-{{ .Cluster }}-test"
    c2:
        name: devcluster`,
			config: FileConfig{
				Defaults: Config{
					DiscoveryImpl: "consul",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"consul": {
							"vtgate-datacenter-tmpl": "dev-{{ .Cluster }}",
							"vtgate-service-name":    "vtgate-svc",
							"vtgate-pool-tag":        "type",
							"vtgate-cell-tag":        "zone",
							"vtgate-addr-tmpl":       "{{ .Name }}.example.com:15999",
						},
					},
				},
				Clusters: map[string]Config{
					"c1": {
						ID:   "c1",
						Name: "testcluster1",
						DiscoveryFlagsByImpl: map[string]map[string]string{
							"consul": {
								"vtgate-datacenter-tmpl": "dev-{{ .Cluster }}-test",
							},
						},
						VtSQLFlags: map[string]string{},
					},
					"c2": {
						ID:                   "c2",
						Name:                 "devcluster",
						DiscoveryFlagsByImpl: map[string]map[string]string{},
						VtSQLFlags:           map[string]string{},
					},
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := FileConfig{
				Defaults: Config{
					DiscoveryFlagsByImpl: map[string]map[string]string{},
				},
				Clusters: map[string]Config{},
			}

			err := yaml.Unmarshal([]byte(tt.yaml), &cfg)
			if tt.err != nil {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.config, cfg)
		})
	}
}

func TestCombine(t *testing.T) {
	tests := []struct {
		name     string
		fc       FileConfig
		defaults Config
		configs  map[string]Config
		expected []Config
	}{
		{
			name: "default overrides file",
			fc: FileConfig{
				Defaults: Config{
					DiscoveryImpl: "consul",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"consul": {
							"vtgate-datacenter-tmpl": "dev-{{ .Cluster }}",
						},
					},
				},
			},
			defaults: Config{
				DiscoveryImpl:        "zk",
				DiscoveryFlagsByImpl: map[string]map[string]string{},
			},
			configs: map[string]Config{
				"1": {
					ID:   "1",
					Name: "one",
				},
				"2": {
					ID:   "2",
					Name: "two",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"consul": {
							"vtgate-datacenter-tmpl": "dev-{{ .Cluster }}-test",
						},
					},
				},
			},
			expected: []Config{
				{
					ID:            "1",
					Name:          "one",
					DiscoveryImpl: "zk",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"consul": {
							"vtgate-datacenter-tmpl": "dev-{{ .Cluster }}",
						},
					},
					VtSQLFlags: map[string]string{},
				},
				{
					ID:            "2",
					Name:          "two",
					DiscoveryImpl: "zk",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"consul": {
							"vtgate-datacenter-tmpl": "dev-{{ .Cluster }}-test",
						},
					},
					VtSQLFlags: map[string]string{},
				},
			},
		},
		{
			name: "mixed",
			fc: FileConfig{
				Defaults: Config{
					DiscoveryImpl: "consul",
				},
				Clusters: map[string]Config{
					"c1": {
						ID:   "c1",
						Name: "cluster1",
					},
					"c2": {
						ID:   "c2",
						Name: "cluster2",
					},
				},
			},
			defaults: Config{
				DiscoveryFlagsByImpl: map[string]map[string]string{
					"zk": {
						"flag": "val",
					},
				},
			},
			configs: map[string]Config{
				"c1": {
					ID:   "c1",
					Name: "cluster1",
				},
				"c3": {
					ID:   "c3",
					Name: "cluster3",
				},
			},
			expected: []Config{
				{
					ID:            "c1",
					Name:          "cluster1",
					DiscoveryImpl: "consul",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"zk": {
							"flag": "val",
						},
					},
					VtSQLFlags: map[string]string{},
				},
				{
					ID:            "c2",
					Name:          "cluster2",
					DiscoveryImpl: "consul",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"zk": {
							"flag": "val",
						},
					},
					VtSQLFlags: map[string]string{},
				},
				{
					ID:            "c3",
					Name:          "cluster3",
					DiscoveryImpl: "consul",
					DiscoveryFlagsByImpl: map[string]map[string]string{
						"zk": {
							"flag": "val",
						},
					},
					VtSQLFlags: map[string]string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.fc.Combine(tt.defaults, tt.configs)
			assert.ElementsMatch(t, tt.expected, actual)
		})
	}
}
