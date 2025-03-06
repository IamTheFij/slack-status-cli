package main

import (
	"testing"
	"time"
)

func TestReadDurationArgs(t *testing.T) {
	cases := []struct {
		name             string
		args             []string
		expectedDuration time.Duration
		resultArgs       []string
	}{
		{
			name:             "No args",
			args:             []string{},
			expectedDuration: 0,
			resultArgs:       []string{},
		},
		{
			name:             "No durations",
			args:             []string{"foo", "bar"},
			expectedDuration: 0,
			resultArgs:       []string{"foo", "bar"},
		},
		{
			name:             "duration at beginning",
			args:             []string{"5m", "bar"},
			expectedDuration: time.Minute * 5,
			resultArgs:       []string{"bar"},
		},
		{
			name:             "duration in the middle",
			args:             []string{"foo", "5m", "bar"},
			expectedDuration: 0,
			resultArgs:       []string{"foo", "5m", "bar"},
		},
		{
			name:             "duration at end without prefix",
			args:             []string{"foo", "bar", "5m"},
			expectedDuration: 0,
			resultArgs:       []string{"foo", "bar", "5m"},
		},
		{
			name:             "duration at end with prefix",
			args:             []string{"foo", "for", "5m"},
			expectedDuration: time.Minute * 5,
			resultArgs:       []string{"foo"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actualArgs, actualDuration := readDurationArgs(c.args)

			if len(actualArgs) != len(c.resultArgs) {
				t.Errorf("expected args %v actual %v", c.resultArgs, actualArgs)
			}

			for i, expected := range c.resultArgs {
				if expected != actualArgs[i] {
					t.Errorf("expected arg[%d] %v actual %v", i, expected, actualArgs[i])
				}
			}

			if actualDuration == nil {
				if c.expectedDuration != 0 {
					t.Errorf("expected %v actual %v", c.expectedDuration, actualDuration)
				}
			} else {
				if *actualDuration != c.expectedDuration {
					t.Errorf("expected %v actual %v", c.expectedDuration, actualDuration)
				}
			}
		})
	}
}
