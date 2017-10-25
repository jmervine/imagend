### `imagend` an (ima)ge (gen)erator for (d)docker

Generate many docker images from a manifest, using templates.

> **This is a prototype and shouldn't be used by anyome but me right now.
> it's currently not very flexible, and probably won't support your project.
> If you're up to it, though, I'm happy to review PRs.**

```
# w/ make
make
make install

# w/o make
go get -v github.com/kardianos/govendor
govendor test -v +local
govendor install +local
```

See manifest.yml.sample and templates/Dockerfile.sample.tmpl for examples.

