package main

import (
	"log"
	"os"
	"path/filepath"
)

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		log.Panic("--- ERROR " + err.Error())
	}

	return true
}

// expand returns the filepath.Abs of a string, optionally running mkdir on it
// ... currently fails if the path isn't a directory ... should probably address
// that at somepoint.
func expand(path string, mkdir bool) string {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	if !exists(path) {
		if !mkdir {
			log.Fatal("--- ERROR file or directory not found: ", path)
		}
		if err := os.Mkdir(path, 0755); err != nil {
			log.Fatal(err)
		}
	}

	return path
}
