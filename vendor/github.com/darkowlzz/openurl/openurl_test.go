package openurl

import (
	"reflect"
	"testing"
)

func TestGetCmd(t *testing.T) {
	cases := []struct {
		platform    string
		expectedCmd string
		expectedErr error
	}{
		{
			platform:    "windows",
			expectedCmd: "explorer.exe",
			expectedErr: nil,
		},
		{
			platform:    "darwin",
			expectedCmd: "open",
			expectedErr: nil,
		},
		{
			platform:    "linux",
			expectedCmd: "xdg-open",
			expectedErr: nil,
		},
		{
			platform:    "freebsd",
			expectedCmd: "",
			expectedErr: unsupportedPlatformError,
		},
	}

	for _, c := range cases {
		cmd, err := getCmd(c.platform)
		if cmd != c.expectedCmd {
			t.Fatalf("Expected cmd to be %s, got %s", c.expectedCmd, cmd)
		}
		if !reflect.DeepEqual(err, c.expectedErr) {
			t.Fatalf("Expected error to be %v, got %v", c.expectedErr, err)
		}
	}
}

func TestOpen(t *testing.T) {
	cases := []struct {
		url         string
		expectedErr error
	}{
		{
			url:         "https://google.com",
			expectedErr: nil,
		},
		{
			url:         "https://example.com",
			expectedErr: nil,
		},
		{
			url:         "",
			expectedErr: invalidUrlError,
		},
	}

	for _, c := range cases {
		err := Open(c.url)
		if !reflect.DeepEqual(err, c.expectedErr) {
			t.Fatalf("Expected error to be %v, got %v", c.expectedErr, err)
		}
	}

	// Test for error returned from getCmd
	getCmd = func(platform string) (string, error) {
		return "", unsupportedPlatformError
	}
	err := Open("https://example.com")
	if !reflect.DeepEqual(err, unsupportedPlatformError) {
		t.Fatalf("Expected error to be %v, got %v", unsupportedPlatformError, err)
	}
}
