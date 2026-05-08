package step

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPlatform_Apple_Xcodeproj(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "MyApp.xcodeproj"), 0755))

	assert.Equal(t, PlatformApple, detectPlatform(dir))
}

func TestDetectPlatform_Apple_Xcworkspace(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "MyApp.xcworkspace"), 0755))

	assert.Equal(t, PlatformApple, detectPlatform(dir))
}

func TestDetectPlatform_Apple_NestedSubdir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "ios", "MyApp.xcodeproj"), 0755))

	assert.Equal(t, PlatformApple, detectPlatform(dir))
}

func TestDetectPlatform_Android_Gradle(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "build.gradle"), []byte(""), 0644))

	assert.Equal(t, PlatformAndroid, detectPlatform(dir))
}

func TestDetectPlatform_Android_GradleKts(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte(""), 0644))

	assert.Equal(t, PlatformAndroid, detectPlatform(dir))
}

func TestDetectPlatform_Android_NestedSubdir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "android"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "android", "build.gradle"), []byte(""), 0644))

	assert.Equal(t, PlatformAndroid, detectPlatform(dir))
}

func TestDetectPlatform_BothAppleAndAndroid(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "MyApp.xcodeproj"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "build.gradle"), []byte(""), 0644))

	assert.Equal(t, PlatformOther, detectPlatform(dir), "both markers → other")
}

func TestDetectPlatform_EmptyDir(t *testing.T) {
	assert.Equal(t, PlatformOther, detectPlatform(t.TempDir()))
}
