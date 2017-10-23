# `imagend (ima)ge (gen)erator for (d)docker`

Generate many docker images from a manifest, using templates.

> **This is a prototype and shouldn't be used by anyome but me right now.
> it's currently not very flexible, and probably won't support your project.
> If you're up to it, though, I'm happy to review PRs.**

```
go run *.go -h
```

See manifest.yml.sample and templates/Dockerfile.sample.tmpl for examples.

##### TODO

- Oh yeah, don't forget tests... some tests might be good if this is
  going to be portable in any way.
- Support tagging minor and major version, e.g. node:8.x and node:8.
  Could probably split on the "." and then remove one, tag, repeat until
  the array is empty.
- Use `github.com/urfave/cli` for CLI over `flag` built in.
