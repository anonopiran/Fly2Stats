package config_test

import (
	config "Fly2Stats/Config"
	"fmt"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

type pathTypeTestCase struct {
	path string
	err  bool
}

func TestPathType(t *testing.T) {
	cases := []pathTypeTestCase{
		{"test-dir", false},
		{":test-dir", true},
	}
	t.Cleanup(func() {
		for _, tc := range cases {
			os.RemoveAll(tc.path)
		}
	})
	for _, tc := range cases {
		t.Run(fmt.Sprintf("path: %s", tc.path), func(t *testing.T) {
			x := new(config.PathType)
			e := x.SetValue(tc.path)
			if e == nil && tc.err {
				t.Errorf("Expected error, but didn't get")
				return
			} else if e != nil && !tc.err {
				t.Errorf("unexpected error %s", e.Error())
				return
			} else if !tc.err {
				_, err := os.ReadDir(tc.path)
				if err != nil {
					t.Errorf("error %s while getting created dir", err.Error())
					return
				}
			}
		})
	}
}

type logLevelTypeTestCase struct {
	val string
	err bool
}

func TestLogLevelType(t *testing.T) {
	cases := []logLevelTypeTestCase{
		{"DEBUG", false},
		{"INFO", false},
		{"WARNING", false},
		{"ERROR", false},
		{"debug", false},
		{"info", false},
		{"warning", false},
		{"error", false},
		{"asdf", true},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("val: %s", tc.val), func(t *testing.T) {
			x := new(config.LogLevelType)
			e := x.SetValue(tc.val)
			if e == nil && tc.err {
				t.Errorf("Expected error, but didn't get")
				return
			} else if e != nil && !tc.err {
				t.Errorf("unexpected error %s", e.Error())
				return
			} else if !tc.err {
				lv := log.GetLevel()
				if lv.String() != strings.ToLower(tc.val) {
					t.Errorf("log level is not set (%v)", lv)
					return
				}
			}
		})
	}
}
