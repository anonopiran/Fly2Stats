package config_test

import (
	"testing"

	config "github.com/anonopiran/Fly2Stats/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestRabbitUrlType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid input",
			input:       "amqp://user:pass@rbt.local:5672/fly2stats",
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: false,
		},
		{
			name:        "no amqp input",
			input:       "http://localhost:5672",
			expectError: true,
		},
		{
			name:        "invalid url",
			input:       "invalid-url",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result config.RabbitUrlType
			err := result.UnmarshalText([]byte(test.input))
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInfluxUrlType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid input",
			input:       "http://user:pass@influxdb.local:8086/user_stats",
			expectError: false,
		},
		{
			name:        "valid https input",
			input:       "http://user:pass@influxdb.local:8086/user_stats",
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: false,
		},
		{
			name:        "invalid url",
			input:       "invalid-url",
			expectError: true,
		},
	}

	for _, test := range tests {
		var result config.InfluxUrlType
		t.Run(test.name, func(t *testing.T) {
			err := result.UnmarshalText([]byte(test.input))
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTagMapTypeType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult config.TagMapType
		expectError    bool
	}{
		{
			name:           "valid single input",
			input:          "t1:v1",
			expectedResult: config.TagMapType{"t1": "v1"},
			expectError:    false,
		},
		{
			name:           "valid multiple input",
			input:          "t1:v1,t2:v2",
			expectedResult: config.TagMapType{"t1": "v1", "t2": "v2"},
			expectError:    false,
		}, {
			name:           "valid multiple input 2",
			input:          "t1:v1,t2:",
			expectedResult: config.TagMapType{"t1": "v1", "t2": ""},
			expectError:    false,
		}, {
			name:           "empty input",
			input:          "",
			expectedResult: nil,
			expectError:    false,
		},
		{
			name:           "invalid input 1",
			input:          ",",
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:           "invalid input 2",
			input:          "test",
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:           "invalid input 3",
			input:          "test:tt,",
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, test := range tests {
		var result config.TagMapType
		t.Run(test.name, func(t *testing.T) {
			err := result.UnmarshalText([]byte(test.input))
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}
