/*
Package openurl provides a common method for opening URLs in default web
browser across different platforms (Windows, macOS, GNU/Linux).
*/

package openurl

import (
	"errors"
	"os/exec"
	"runtime"
)

var unsupportedPlatformError = errors.New("Unsupported platform")
var invalidUrlError = errors.New("Invalid url")

// Open executes platform specific command to open default web browser with
// the given url
func Open(url string) error {
	if len(url) == 0 {
		return invalidUrlError
	}
	cmd, err := getCmd(runtime.GOOS)
	if err != nil {
		return err
	}
	return exec.Command(cmd, url).Start()
}

// getCmd returns the command that is used to open default web browser,
// given a url
var getCmd = func(platform string) (string, error) {
	switch platform {
	case "windows":
		return "explorer.exe", nil
	case "darwin":
		return "open", nil
	case "linux":
		return "xdg-open", nil
	default:
		return "", unsupportedPlatformError
	}
}
