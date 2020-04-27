package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStorageKey(t *testing.T) {

	t.Run("parseStorageKeyVersion", func(t *testing.T) {

		t.Run("positive cases", func(t *testing.T) {
			tests := []struct {
				in  string
				out *storageKeyVersion
			}{
				{"key0_20060102T150405", &storageKeyVersion{"key0", time.Date(2006, time.Month(1), 2, 15, 4, 5, 0, time.UTC)}},
				{"key1_20180901T120000", &storageKeyVersion{"key1", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"key2_blah_20180901T120000", &storageKeyVersion{"key2_blah", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"key3_blah_blah_20180901T120000", &storageKeyVersion{"key3_blah_blah", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"RP-100_20180901T120000", &storageKeyVersion{"RP-100", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"RP-100_blah_20180901T120000", &storageKeyVersion{"RP-100_blah", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"_key4_20180901T120000", &storageKeyVersion{"_key4", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"key5__20180901T120000", &storageKeyVersion{"key5_", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
				{"hotfix/key6_20180901T120000", &storageKeyVersion{"hotfix_key6", time.Date(2018, time.Month(9), 1, 12, 0, 0, 0, time.UTC)}},
			}

			for _, test := range tests {
				got, err := parseStorageKeyVersion(test.in)
				assert.Nil(t, err)
				assert.Equal(t, test.out, got)
			}
		})

		t.Run("negative cases", func(t *testing.T) {
			tests := []string{
				"",
				"blah",
				"blah_blah",
				"key020180901T120000",
				"key0_20180901T12",
				"key0_20180901T1200000",
				"_20180901T1200000",
				"key1_20060102T150405_",
			}

			for _, test := range tests {
				got, err := parseStorageKeyVersion(test)
				assert.Nil(t, got)
				assert.NotNil(t, err)
			}
		})
	})

	t.Run("parseStorageKey", func(t *testing.T) {

		t.Run("positive cases", func(t *testing.T) {
			tests := []struct {
				in  string
				out *storageKey
			}{
				{
					in: "group0/project0/key0_20060102T150405",
					out: &storageKey{
						"group0",
						"project0",
						&storageKeyVersion{"key0", time.Date(2006, time.Month(1), 2, 15, 4, 5, 0, time.UTC)},
					},
				},
				{
					in: "group0/project0/hotfix/key1_20060102T150405",
					out: &storageKey{
						"group0",
						"project0",
						&storageKeyVersion{"hotfix_key1", time.Date(2006, time.Month(1), 2, 15, 4, 5, 0, time.UTC)},
					},
				},
			}

			for _, test := range tests {
				got, err := parseStorageKey(test.in)
				assert.Nil(t, err)
				assert.Equal(t, test.out, got)
			}
		})
	})

	t.Run("newStorageKey", func(t *testing.T) {

		t.Run("positive cases", func(t *testing.T) {
			tests := []struct {
				group, project, versionKey string
				out                        *storageKey
			}{
				{
					group: "group0", project: "project0", versionKey: "key0",
					out: &storageKey{
						"group0",
						"project0",
						&storageKeyVersion{"key0", time.Time{}},
					},
				},
			}

			for _, test := range tests {
				got := newStorageKey(test.group, test.project, &storageKeyVersion{test.versionKey, time.Time{}})
				assert.Equal(t, test.out, got)
			}
		})
	})
}
