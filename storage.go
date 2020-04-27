package main

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/otiai10/copy"
)

type storage struct {
	basePath string
}

func newStorage(basePath string) (*storage, error) {
	basePath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	return &storage{basePath: basePath}, nil
}

func (s *storage) getStorageKeyPath(key *storageKey) string {
	return filepath.Join(s.basePath, key.getPath())
}

func (s *storage) getResultsPath(key *storageKey) string {
	return filepath.Join(s.getStorageKeyPath(key), "results")
}

func (s *storage) getReportPath(key *storageKey) string {
	return filepath.Join(s.getStorageKeyPath(key), "report")
}

func (s *storage) getProjectPath(key *storageKey) string {
	return filepath.Join(s.getStorageKeyPath(key), "..")
}

func (s *storage) putResults(r *zip.Reader, key *storageKey) error {
	basePath := s.getResultsPath(key)

	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return err
	}

	if r == nil {
		return nil
	}

	for _, f := range r.File {
		filePath := filepath.Join(basePath, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, 0755)
			if err != nil {
				return err
			}

			continue
		}

		dst, err := os.Create(filePath)
		if err != nil {
			return err
		}
		//noinspection GoUnhandledErrorResult,GoDeferInLoop
		defer dst.Close()

		src, err := f.Open()
		if err != nil {
			return err
		}
		//noinspection GoUnhandledErrorResult,GoDeferInLoop
		defer src.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *storage) generateReport(key *storageKey) error {
	resultsPath := s.getResultsPath(key)
	reportPath := s.getReportPath(key)

	log.Printf("generating report from %s to %s", resultsPath, reportPath)
	return exec.Command("allure", "generate", "--output", reportPath, resultsPath, "--clean").Run()
}

func (s *storage) cleaning(interval time.Duration, lifeTime time.Duration) {
	for {
		err := s.deleteBefore(time.Now().Add(-lifeTime))
		if err != nil {
			log.Printf("deleteBefore error: %s", err)
		}

		time.Sleep(interval)
	}
}

func (s *storage) walk(fn func(sk *storageKey) error) error {
	return filepath.Walk(s.basePath, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// check if `fi` is symlink
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		if !fi.IsDir() {
			return nil
		}

		kp := strings.Trim(strings.TrimPrefix(fp, s.basePath), string(os.PathSeparator))

		if strings.Count(kp, string(os.PathSeparator)) != 2 {
			return nil
		}

		sk, err := parseStorageKey(kp)
		if err != nil {
			return err
		}

		return fn(sk)
	})
}

func (s *storage) deleteBefore(ts time.Time) error {
	var keys []*storageKey

	err := s.walk(func(sk *storageKey) error {
		if sk.version.ts.Before(ts) {
			keys = append(keys, sk)
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, key := range keys {
		if err = s.delete(key); err != nil {
			return err
		}
	}

	return nil
}

func (s *storage) delete(sk *storageKey) error {
	return os.RemoveAll(s.getStorageKeyPath(sk))
}

func (s *storage) exist(key *storageKey) bool {
	_, err := os.Stat(s.getStorageKeyPath(key))
	return !os.IsNotExist(err)
}

func (s *storage) copyHistory(key *storageKey) error {
	keys, err := s.listProject(key)
	if err != nil {
		return err
	}

	if len(keys) <= 1 {
		return nil
	}

	// sort keys by timestamp (last reports first)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].version.ts.After(keys[j].version.ts)
	})

	prev := filepath.Join(s.getReportPath(keys[1]), "history")
	// if history folder is not exists
	_, err = os.Stat(prev)
	if os.IsNotExist(err) {
		return nil
	}

	dst := filepath.Join(s.getResultsPath(keys[0]), "history")

	log.Printf("copy history from %s to %s", prev, dst)
	err = copy.Copy(prev, dst)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) listProject(key *storageKey) ([]*storageKey, error) {
	var keys []*storageKey

	// extract all storage keys for project
	files, err := ioutil.ReadDir(s.getProjectPath(key))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.Name() != "last" {
			k, err := parseStorageKey(filepath.Join(key.group, key.project, f.Name()))
			if err != nil {
				return nil, err
			}
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (s *storage) createLastVersionSymlink(key *storageKey) error {
	symlink := filepath.Join(s.getProjectPath(key), "last")

	// remove previous existent symlink
	_, err := os.Lstat(symlink)
	if err == nil {
		err = os.Remove(symlink)
		if err != nil {
			return err
		}
	}

	// create symlink
	target := s.getStorageKeyPath(key)
	err = os.Symlink(target, symlink)
	if err != nil {
		return err
	}
	return nil
}
