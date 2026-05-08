package step

import (
	"os"
	"path/filepath"
	"strings"
)

type Platform string

const (
	PlatformApple   Platform = "apple"
	PlatformAndroid Platform = "android"
	PlatformOther   Platform = "other"
)

// detectPlatform scans dir for project markers to determine the platform.
// Checks both the root directory and one level deep (e.g. ios/, android/).
func detectPlatform(dir string) Platform {
	hasApple := false
	hasAndroid := false

	scan := func(entries []os.DirEntry) {
		for _, e := range entries {
			name := e.Name()
			if strings.HasSuffix(name, ".xcodeproj") || strings.HasSuffix(name, ".xcworkspace") {
				hasApple = true
			}
			if name == "build.gradle" || name == "build.gradle.kts" {
				hasAndroid = true
			}
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return PlatformOther
	}
	scan(entries)

	// Check one level deep (e.g. ios/MyApp.xcodeproj, android/build.gradle)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subEntries, err := os.ReadDir(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		scan(subEntries)
	}

	if hasApple && hasAndroid {
		return PlatformOther
	}
	if hasApple {
		return PlatformApple
	}
	if hasAndroid {
		return PlatformAndroid
	}
	return PlatformOther
}
