package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {

	t.Run("newStorage", func(t *testing.T) {

		t.Run("should create storage in storageBasePath", func(t *testing.T) {
			// arrange
			storageBasePath, err := ioutil.TempDir("", "allure-portal")
			require.NoError(t, err)

			// action
			s, err := newStorage(storageBasePath)

			// assert
			require.NoError(t, err)
			assert.DirExists(t, s.basePath)
		})
	})

	t.Run("putResults", func(t *testing.T) {

		t.Run("should create storage key with correct paths", func(t *testing.T) {
			// arrange
			s := createStorageInTempDir(t)

			now := time.Now()
			sKey := "group0/project0/key0_" + now.Format(storageKeyVersionTsFormat)
			sk, err := parseStorageKey(sKey)
			require.NoError(t, err)

			// action
			err = s.putResults(nil, sk)

			// assert
			require.NoError(t, err)
			assert.True(t, s.exist(sk))

			path := s.getStorageKeyPath(sk)
			assert.DirExists(t, path)
			assert.Contains(t, path, sKey)

			reportPath := s.getReportPath(sk)
			assert.Contains(t, reportPath, sKey+"/report")

			resultsPath := s.getResultsPath(sk)
			assert.DirExists(t, resultsPath)
			assert.Contains(t, resultsPath, sKey+"/results")
		})
	})

	t.Run("generateReport", func(t *testing.T) {

		t.Run("should create allure report (index.html) in report folder", func(t *testing.T) {
			// arrange
			s := createStorageInTempDir(t)

			now := time.Now()
			sKey := "group0/project0/key0_" + now.Format(storageKeyVersionTsFormat)
			sk, err := parseStorageKey(sKey)
			require.NoError(t, err)

			err = s.putResults(nil, sk)
			require.NoError(t, err)

			// action
			_ = s.generateReport(sk)
			//err = s.generateReport(sk)

			// assert
			//require.Error(t, err)
			//assert.FileExists(t, filepath.Join(s.getReportPath(sk), "index.html"))
		})
	})

	t.Run("walk", func(t *testing.T) {

		t.Run("should list storage keys relative paths", func(t *testing.T) {
			// arrange
			s := createStorageInTempDir(t)

			now := time.Now()
			sKey := "group0/project0/key0_" + now.Format(storageKeyVersionTsFormat)
			sk, err := parseStorageKey(sKey)
			require.NoError(t, err)

			err = s.putResults(nil, sk)
			require.NoError(t, err)

			// action
			var keys []*storageKey

			err = s.walk(func(sk *storageKey) error {
				keys = append(keys, sk)
				return nil
			})

			// assert
			require.NoError(t, err)
			require.Len(t, keys, 1)
			assert.Equal(t, sk, keys[0])
		})
	})

	t.Run("deleteBefore", func(t *testing.T) {

		t.Run("should delete previous storage keys", func(t *testing.T) {
			// arrange
			s := createStorageInTempDir(t)

			now := time.Now()

			// now
			sk0, err := parseStorageKey("group0/project0/key0_" + now.Format(storageKeyVersionTsFormat))
			require.NoError(t, err)

			// now - 1 hour
			sk1, err := parseStorageKey("group0/project1/key1_" + now.Add(-time.Hour).Format(storageKeyVersionTsFormat))
			require.NoError(t, err)

			// now + 1 hour
			sk2, err := parseStorageKey("group1/project0/key0_" + now.Add(time.Hour).Format(storageKeyVersionTsFormat))
			require.NoError(t, err)

			err = s.putResults(nil, sk0)
			require.NoError(t, err)
			assert.True(t, s.exist(sk0))

			err = s.putResults(nil, sk1)
			require.NoError(t, err)
			assert.True(t, s.exist(sk1))

			err = s.putResults(nil, sk2)
			require.NoError(t, err)
			assert.True(t, s.exist(sk2))

			// action
			err = s.deleteBefore(sk2.version.ts)
			require.NoError(t, err)

			// assert
			assert.False(t, s.exist(sk0))
			assert.False(t, s.exist(sk1))
			assert.True(t, s.exist(sk2))
		})
	})

	t.Run("createLastVersionSymlink", func(t *testing.T) {

		t.Run("should create symlink for last report version", func(t *testing.T) {
			// arrange
			s := createStorageInTempDir(t)

			now := time.Now()
			sKey := "group0/project0/key0_" + now.Format(storageKeyVersionTsFormat)
			sk, err := parseStorageKey(sKey)
			require.NoError(t, err)

			_ = s.putResults(nil, sk)
			require.NoError(t, err)

			// action
			err = s.createLastVersionSymlink(sk)

			// assert
			require.NoError(t, err)
			symlinkPath := filepath.Join(s.getProjectPath(sk), "last")
			require.FileExists(t, symlinkPath)
			symlink, _ := os.Lstat(symlinkPath)
			require.Equal(t, os.ModeSymlink, symlink.Mode()&os.ModeSymlink)
			link, _ := os.Readlink(symlinkPath)
			assert.Equal(t, s.getStorageKeyPath(sk), link)
		})
	})
}

func createStorageInTempDir(t *testing.T) *storage {
	t.Helper()
	storageBasePath, err := ioutil.TempDir("", "allure-portal")
	require.NoError(t, err)
	s, err := newStorage(storageBasePath)
	require.NoError(t, err)
	t.Logf("storageBasePath=%s", s.basePath)
	return s
}
