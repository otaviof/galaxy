package galaxy

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// fatal replacement for panic during runtime, no strack-trace.
func fatal(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	os.Exit(1)
}

// fileExists Check if path exists, boolean return.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// readFile Wrap up a ioutil call, using fatal log in case of error.
func readFile(path string) []byte {
	if !fileExists(path) {
		fatal("[ERROR] Can't find file '%s'", path)
	}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		fatal("[ERROR] Reading bytes: '%s'", err)
	}
	return fileBytes
}

// isDir Check if informed path is a directory, boolean return.
func isDir(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

// formatSlice pretty print a string slice using commas.
func formatSlice(slice []string) string {
	return fmt.Sprintf("[%s]", strings.Join(slice, ", "))
}

// stringSliceContains checks if a slice contiains a string.
func stringSliceContains(slice []string, str string) bool {
	var sliceStr string

	for _, sliceStr = range slice {
		if str == sliceStr {
			return true
		}
	}

	return false
}

// SetLogLevel parse and set logrus log-level.
func SetLogLevel(levelStr string) {
	var level log.Level
	var err error

	if level, err = log.ParseLevel(levelStr); err != nil {
		log.Fatalf("[ERROR] Setting log-level ('%s'): %s", levelStr, err)
	}
	log.SetLevel(level)
}
