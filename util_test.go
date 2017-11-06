package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_exists(t *testing.T) {
	f := mkfile()
	defer rmfile(f)

	assert := assert.New(t)
	assert.True(exists(f))
}

func Test_expand(t *testing.T) {
	f := expand("_support/fixtures/manifest.yml", false)

	assert := assert.New(t)
	assert.True(strings.HasSuffix(f, "manifest.yml"))
	assert.True(strings.HasPrefix(f, "/"))
}

func Test_containsAny(t *testing.T) {
	assert := assert.New(t)
	strings := []string{"a", "b", "c"}
	assert.True(containsAny(strings, "x", "c"))
	assert.True(containsAny(strings, "c", "x"))
	assert.False(containsAny(strings, "x"))
}

func mkfile() string {
	tmpfile, err := ioutil.TempFile("", "util_test")
	if err != nil {
		panic(err)
	}

	defer tmpfile.Close()
	return tmpfile.Name()
}

func rmfile(path string) {
	os.Remove(path)
}
