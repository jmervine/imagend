# imagend
Generate many docker images from a manifest, using templates.

```
go run *.go -h
```

See manifest.yml.sample and templates/Dockerfile.sample.tmpl for examples.

##### TODO

- Because the current alpha sorting won't always be good enough, add the
  concept of dependencies. The basic idea is to have an array that holds
  order and another for storing images that have been built. Then I
  should be able to iterate over versions, as they come in, building any
  dependencies found, marking them as built and then skipping them
  once it's their natural turn.
- Support tagging minor and major version, e.g. node:8.x and node:8.
  Could probably split on the "." and then remove one, tag, repeat until
  the array is empty.
- Add alias support, e.g. jmervine/herokudev-base -> jmervine/herokudev.
  This should also support all versions including latest.
- Extrat generator in to it's own repo for use with other projects, e.g.
  jmervine/mini\*
- Consider supporting `herokudev-rails` again.
