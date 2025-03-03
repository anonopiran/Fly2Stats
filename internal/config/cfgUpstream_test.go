package config_test

import (
	"net/url"
	"testing"

	config "github.com/anonopiran/Fly2Stats/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamUrlType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectedErr bool
		output      *config.UpstreamUrlType
	}{
		{
			// Happy path with valid GRPC URL and server type

			name:        "valid_grpc_url",
			input:       []byte("grpc://v2fly@test.server:8080"),
			expectedErr: false,
			output: &config.UpstreamUrlType{
				URL: url.URL{
					Scheme: "grpc",
					Host:   "test.server:8080",
				},
				ServerType: config.V2FLY_SRV,
			},
		}, {
			name:        "valid_grpc_url",
			input:       []byte("grpc://xray@test.server:8080"),
			expectedErr: false,
			output: &config.UpstreamUrlType{
				URL: url.URL{
					Scheme: "grpc",
					Host:   "test.server:8080",
				},
				ServerType: config.XRAY_SRV,
			},
		},
		// Invalid server type
		{
			name:        "valid_setver_type",
			input:       []byte("grpc://zzz@test.server:8080"),
			expectedErr: true,
			output:      &config.UpstreamUrlType{},
		},
		// Invalid scheme
		{
			name:        "invalid_scheme",
			input:       []byte("http://server.example.com:8080/V2FLY_SRV"),
			expectedErr: true,
			output:      &config.UpstreamUrlType{},
		},
		// Missing hostname
		{
			name:        "missing_hostname",
			input:       []byte("grpc:///V2FLY_SRV"),
			expectedErr: true,
			output:      &config.UpstreamUrlType{},
		},
		// Missing port
		{
			name:        "missing_port",
			input:       []byte("grpc://server.example.com/V2FLY_SRV"),
			expectedErr: true,
			output:      &config.UpstreamUrlType{},
		},
		// Unsupported server type
		{
			name:        "unsupported_servertype",
			input:       []byte("grpc://server.example.com:8080/UNSUPPORTED_SRV"),
			expectedErr: true,
			output:      &config.UpstreamUrlType{},
		}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var f config.UpstreamUrlType
			err := f.UnmarshalText(tc.input)

			// Assertions using testify
			if tc.expectedErr {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tc.output, &f)
		})
	}
}
