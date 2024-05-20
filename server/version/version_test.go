package version

import (
	"testing"
)

func TestConvertToTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected int64
		wantErr  bool
	}{
		{
			name:     "ValidDate",
			date:     "2024-05-19T13:17:37",
			expected: 1716124657,
			wantErr:  false,
		},
		{
			name:     "InvalidDate",
			date:     "invalid",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToTimestamp(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("convertToTimestamp() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildVersion(t *testing.T) {
	Version = "test_version"
	CommitHash = "test_commit"
	BuildTimestamp = "2024-05-19T13:17:37"

	wantVersion := "test_commit"
	wantTimestamp := int64(1716124657)

	version, timestamp, _ := BuildVersion()

	if version != wantVersion {
		t.Errorf("BuildVersion() version = %v, want %v", version, wantVersion)
	}
	if timestamp != wantTimestamp {
		t.Errorf("BuildVersion() timestamp = %v, want %v", timestamp, wantTimestamp)
	}
}
