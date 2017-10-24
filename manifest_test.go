package main

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

func init() {
	reset()
}

func Test_loadManifest(t *testing.T) {
	assert := assert.New(t)

	m := loadManifest(manifile)
	assert.Equal(2, len(m))
	assert.Equal("one", m[0].Name)
	assert.Equal("two", m[1].Name)

	monkey.Patch(log.Fatal, func(a ...interface{}) {
		assert.Equal("open some/bad/file/path.yml: no such file or directory", a[0].(error).Error())
	})
	defer monkey.Unpatch(log.Fatal)

	loadManifest("some/bad/file/path.yml")
}

func TestManifest_builds_all(t *testing.T) {
	assert := assert.New(t)
	m := loadFixture()

	assert.Equal(m, m.builds())
}

func TestManifest_builds_with_image(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(log.Fatal, func(a ...interface{}) {
		assert.Fail("log.Fatal shouldn't have been called")
	})
	defer monkey.Unpatch(log.Fatal)

	m := loadFixture()

	image = "one"
	defer reset()

	b := m.builds()

	assert.NotEqual(m, b)
	assert.Equal(1, len(b))
	assert.Equal("one", b[0].Name)
}

func TestManifest_builds_with_invalid_image(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(log.Fatal, func(a ...interface{}) {
		assert.Equal([]interface{}{"ERROR requested image not found: ", "missing"}, a)
	})
	defer monkey.Unpatch(log.Fatal)

	m := loadFixture()

	image = "missing"
	defer reset()

	m.builds()
}

func TestManifest_builds_with_invalid_version(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(log.Fatal, func(a ...interface{}) {
		assert.Equal([]interface{}{"ERROR requested version not found: ", "missing"}, a)
	})
	defer monkey.Unpatch(log.Fatal)

	m := loadFixture()

	image = "one"
	version = "missing"
	defer reset()

	m.builds()
}

func TestVersion_imageName(t *testing.T) {
	assert := assert.New(t)

	v := Version{Name: "name"}
	assert.Equal("name", v.imageName())

	v = Version{Image: "image"}
	assert.Equal("image", v.imageName())
}

func TestVersion_tag(t *testing.T) {
	assert := assert.New(t)

	v := Version{Name: "name", Version: "1"}
	assert.Equal("name:1", v.tag())

	v = Version{Image: "image", Version: "1"}
	assert.Equal("image:1", v.tag())
}

func TestVersion_tags(t *testing.T) {
	assert := assert.New(t)

	// Basic
	v := Version{Name: "name", Version: "1"}
	assert.Equal([]string{"name:1"}, v.tags())

	// Latest
	v = Version{Image: "image", Version: "1", Latest: true}
	assert.Equal([]string{"image:1", "image:latest"}, v.tags())

	// Aliases
	v = Version{Name: "name", Version: "1", Aliases: []string{"my/name"}}
	assert.Equal([]string{"name:1", "my/name:1"}, v.tags())

	// Aliases + Latest
	v = Version{Name: "name", Version: "1", Aliases: []string{"my/name"}, Latest: true}
	assert.Equal([]string{"name:1", "name:latest", "my/name:1", "my/name:latest"}, v.tags())
}

func TestVersion_template(t *testing.T) {
	defer reset()

	assert := assert.New(t)

	v := Version{Name: "test"}
	assert.Contains(v.template(), "Dockerfile.test.tmpl")

	monkey.Patch(log.Fatal, func(a ...interface{}) {
		assert.Equal("ERROR template file not found: ", a[0].(string))
		assert.Contains(a[1].(string), "Dockerfile.missing.tmpl")
	})
	defer monkey.Unpatch(log.Fatal)

	v = Version{Name: "missing"}
	v.template()
}

func TestVersion_dockerbase(t *testing.T) {
	assert := assert.New(t)

	monkey.Patch(log.Fatal, func(a ...interface{}) {
		assert.Fail("log.Fatal shouldn't have been called")
	})
	defer monkey.Unpatch(log.Fatal)

	v := Version{Name: "test", Version: "1"}
	db := v.dockerbase()

	assert.Contains(db, "test-1")
}

func loadFixture() Manifest {
	return loadManifest(manifile)
}

func reset() {
	// testing defaults
	tmpldir = path.Join("_support", "fixtures")
	manifile = path.Join("_support", "fixtures", "manifest.yml")
	outdir = os.TempDir()
	image = ""
	version = ""
}
