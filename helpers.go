package main

import "strings"

func replaceSlashesToUnderscores(s string) string {
	return strings.Replace(s, "/", "_", -1)
}
