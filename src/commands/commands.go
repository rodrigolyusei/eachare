package commands

import (
	"io/fs"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func GetSharedDirectory(sharedPath string) []fs.DirEntry {
	entries, err := os.ReadDir(sharedPath)
	check(err)

	return entries
}
