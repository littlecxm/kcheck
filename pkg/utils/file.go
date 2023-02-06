package utils

import (
	"os"
	"strings"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	return !os.IsNotExist(err) && !info.IsDir()
}

func CheckPathBlacklist(path string) bool {
	list := []string{
		".ds_store",
		".git",
	}
	for _, s := range list {
		if strings.Contains(strings.ToLower(strings.ToLower(path)), s) {
			return true
		}
	}
	return false
}
