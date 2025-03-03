package config_test

import (
	"net/url"
	"os"
	"testing"

	config "github.com/anonopiran/Fly2Stats/internal/config"

	"github.com/stretchr/testify/assert"
)

func validSampleConfig() config.ConfigType {
	inflxURL := config.InfluxUrlType{}
	rbtURL := config.RabbitUrlType{}
	rayUrl1 := config.UpstreamUrlType{}
	rayUrl2 := config.UpstreamUrlType{}
	inflxURL.UnmarshalText([]byte("http://user:pass@influxdb.local:8086/user_stats"))
	rbtURL.UnmarshalText([]byte("amqp://user:pass@rbt.local:5672/fly2stats"))
	rayUrl1.UnmarshalText([]byte("grpc://v2fly@test.server:8080"))
	rayUrl2.UnmarshalText([]byte("grpc://xray@test.server:8080"))
	return config.ConfigType{
		Supervisor: config.SupervisorConfigType{
			Interval:       60,
			InfluxdbUrl:    inflxURL,
			InfluxdbTags:   map[string]string{"k1": "v1", "k2": "v2"},
			RabbitUrl:      rbtURL,
			RabbitExchange: "exch",
			CheckpointPath: "/my/test/path/",
		},
		Upstream: config.UpstreamConfigType{
			Address: []config.UpstreamUrlType{rayUrl1, rayUrl2},
		},
		LogLevel: "WARNING",
	}
}
func TestLoadDotEnv(t *testing.T) {
	tests := []struct {
		name       string
		envContent string
		envKey     []string
		envValue   []string
	}{
		{
			name:       "No .env file",
			envContent: "",
		},
		{
			name:       ".env file with valid content",
			envContent: "KEY=VALUE\n",
			envKey:     []string{"KEY"},
			envValue:   []string{"VALUE"},
		},
		{
			name:       ".env file with multiple variables",
			envContent: "KEY1=VALUE1\nKEY2=VALUE2\n",
			envKey:     []string{"KEY1", "KEY2"},
			envValue:   []string{"VALUE1", "VALUE2"},
		},
	}
	// ...
	f, err := os.MkdirTemp("", "envtest")
	if err != nil {
		panic(err)
	}
	oldPath := os.Getenv("path")
	os.Setenv("path", f)
	defer func() { os.Setenv("path", oldPath) }()
	// ...
	// ...
	for _, tt := range tests {
		if tt.envContent != "" {
			os.Clearenv()
			file, err := os.Create(".env")
			if err != nil {
				panic(err)
			}
			_, err = file.WriteString(tt.envContent)
			file.Close()
			if err != nil {
				panic(err)
			}
		}
		// ...
		os.Clearenv()
		config.LoadDotEnv()
		if tt.envKey != nil {
			for c, k := range tt.envKey {
				value, exists := os.LookupEnv(k)
				assert.True(t, exists)
				assert.Equal(t, tt.envValue[c], value)
			}

		}
		// ...
		if tt.envContent != "" {
			os.Remove(".env")
		}
	}
}

// ...
func TestDoUnmarshal(t *testing.T) {
	up_urls := []url.URL{}
	for _, u := range []string{"grpc://xray@test.com:123", "grpc://v2fly@test2.com:124"} {
		ur, _ := url.Parse(u)
		up_urls = append(up_urls, *ur)
	}
	inflxUrl := config.InfluxUrlType{}
	inflxUrl.UnmarshalBinary([]byte("http://user:pass@influxdb.local:8086/user_stats"))
	rbtUrl := config.RabbitUrlType{}
	rbtUrl.UnmarshalBinary([]byte("amqp://user:pass@rbt.local:5672/fly2stats"))

	tests := []struct {
		name      string
		envVars   map[string]string
		expected  config.ConfigType
		expectErr bool
	}{
		{
			name:    "Default configuration",
			envVars: map[string]string{},
			expected: config.ConfigType{
				LogLevel: "WARNING",
				Supervisor: config.SupervisorConfigType{
					Interval: 60,
				},
			},
			expectErr: false,
		},
		{
			name: "Custom conf",
			envVars: map[string]string{
				"LOG_LEVEL":                   "DEBUG",
				"SUPERVISOR__INTERVAL":        "80",
				"SUPERVISOR__INFLUXDB_URL":    "http://user:pass@influxdb.local:8086/user_stats",
				"SUPERVISOR__INFLUXDB_TAGS":   "x1:t1,x2:t2",
				"SUPERVISOR__RABBIT_URL":      "amqp://user:pass@rbt.local:5672/fly2stats",
				"SUPERVISOR__RABBIT_EXCHANGE": "exch",
				"SUPERVISOR__CHECKPOINT_PATH": "my/path/test",
				"UPSTREAM__ADDRESS":           "grpc://xray@test.com:123,grpc://v2fly@test2.com:124",
			},
			expected: config.ConfigType{
				LogLevel: "DEBUG",
				Supervisor: config.SupervisorConfigType{
					Interval:       80,
					InfluxdbUrl:    inflxUrl,
					InfluxdbTags:   map[string]string{"x1": "t1", "x2": "t2"},
					RabbitUrl:      rbtUrl,
					RabbitExchange: "exch",
					CheckpointPath: "my/path/test",
				},
				Upstream: config.UpstreamConfigType{
					Address: []config.UpstreamUrlType{{
						URL:        up_urls[0],
						ServerType: config.XRAY_SRV,
					}, {
						URL:        up_urls[1],
						ServerType: config.V2FLY_SRV,
					}},
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			defer func() {
				// Unset environment variables after the test
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			var cfg config.ConfigType
			err := config.DoUnmarshal(&cfg)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.LogLevel, cfg.LogLevel)
				assert.Equal(t, tt.expected.Supervisor.Interval, cfg.Supervisor.Interval)
			}
		})
	}
}

// ...
func TestDoValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    config.ConfigType
		expectErr bool
	}{
		{
			name: "Valid configuration",
			config: func() config.ConfigType {
				c := validSampleConfig()
				return c
			}(),
			expectErr: false,
		}, {
			name: "Missing supervisor",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor = config.SupervisorConfigType{}
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing upstream",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Upstream = config.UpstreamConfigType{}
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing log level",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.LogLevel = ""
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing Upstream.Address",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Upstream.Address = []config.UpstreamUrlType{}
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing Supervisor.CheckpointPath",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor.CheckpointPath = ""
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing Supervisor.InfluxdbTags",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor.InfluxdbTags = map[string]string{}
				return c
			}(),
			expectErr: false,
		}, {
			name: "Missing Supervisor.InfluxdbUrl",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor.InfluxdbUrl = config.InfluxUrlType{}
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing Supervisor.Interval",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor.Interval = 0
				return c
			}(),
			expectErr: true,
		}, {
			name: "Missing Supervisor.RabbitUrl and Supervisor.RabbitExchange",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor.RabbitExchange = ""
				c.Supervisor.RabbitUrl = config.RabbitUrlType{}
				return c
			}(),
			expectErr: false,
		}, {
			name: "Missing Supervisor.RabbitExchange",
			config: func() config.ConfigType {
				c := validSampleConfig()
				c.Supervisor.RabbitExchange = ""
				return c
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.DoValidate(&tt.config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
