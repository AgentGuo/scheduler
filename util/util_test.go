package util

import (
	"testing"
)

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		id     string
		suffix string
		want   string
	}{
		{"test#1", "besteffort-pod", "abc", ".slice", "besteffort-podabc.slice"},
		{"test#2", "burstable-pod", "abc", ".slice", "burstable-podabc.slice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinPath(tt.prefix, tt.id, tt.suffix); got != tt.want {
				t.Errorf("JoinPath(%s, %s, %s) is %s, want %s", tt.prefix, tt.id, tt.suffix, got, tt.want)
			}
		})
	}
}

func TestIsDirOrFileExist(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"test#1", "/root/test", true},
		{"test#2", "/root/test_non", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := IsDirOrFileExist(tt.path); got != tt.want {
				t.Errorf("IsDirOrFileExist(%s) is %t, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestWriteIntToFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		data int64
		want error
	}{
		{"test#1", "/root/test", 20000, nil},
		{"test#2", "/root/test_non", 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err1 := IsDirOrFileExist(tt.path); ok {
				if err := WriteIntToFile(tt.path, tt.data); err != tt.want {
					t.Errorf("WriteIntToFile(%s, %d) is %s, want %v", tt.path, tt.data, err, tt.want)
				}
			} else {
				t.Errorf("error %v", err1)
			}

		})
	}
}
