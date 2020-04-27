package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	unknownKeyVersion         = "unknown"
	storageKeyVersionTsFormat = "20060102T150405"
)

type storageKeyVersion struct {
	key string
	ts  time.Time
}

func newStorageKeyVersion(key string, ts time.Time) *storageKeyVersion {
	return &storageKeyVersion{
		key: replaceSlashesToUnderscores(key),
		ts:  ts,
	}
}

func parseStorageKeyVersion(s string) (*storageKeyVersion, error) {
	p := strings.Split(s, "_")
	if len(p) < 2 {
		return nil, fmt.Errorf("wrong storage key version: %s", s)
	}

	pts := p[len(p)-1]

	ts, err := time.Parse(storageKeyVersionTsFormat, pts)
	if err != nil {
		return nil, fmt.Errorf("wrong storage key version: %s: %s", s, err)
	}

	return &storageKeyVersion{
		key: replaceSlashesToUnderscores(strings.TrimSuffix(s, "_"+pts)),
		ts:  ts,
	}, nil
}

func (v *storageKeyVersion) String() string {
	return fmt.Sprintf("%s_%s", v.key, v.ts.Format(storageKeyVersionTsFormat))
}

type storageKey struct {
	group   string
	project string
	version *storageKeyVersion
}

func newStorageKey(group, project string, version *storageKeyVersion) *storageKey {
	return &storageKey{
		group:   replaceSlashesToUnderscores(group),
		project: replaceSlashesToUnderscores(project),
		version: version,
	}
}

func parseStorageKey(s string) (*storageKey, error) {
	p := strings.SplitN(strings.Trim(s, "/"), "/", 3)
	if len(p) < 3 {
		return nil, fmt.Errorf("wrong storage key: %s", s)
	}

	version, err := parseStorageKeyVersion(p[2])
	if err != nil {
		return nil, fmt.Errorf("wrong storage key: %s: %s", s, err)
	}

	return &storageKey{p[0], p[1], version}, nil
}

func (k *storageKey) getPath() string {
	if k.version == nil {
		k.version = newStorageKeyVersion(unknownKeyVersion, time.Now())
	}

	return filepath.Join(k.group, k.project, k.version.String())
}

func (k *storageKey) String() string {
	return k.getPath()
}
