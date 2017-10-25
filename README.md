### `imagend` an (ima)ge (gen)erator for (d)docker

Generate many docker images from a manifest, using templates.

### Build / Install

```
got get -v -u github.com/jmervine/imagend
```

### Overview

Do you, by chance, maintain a bunch of docker images? With a bunch of version?
All of which need to be rebuilt from time to time, like when there's a CVE that
needs addressing, or dependancy that needs updating? Of course you don't, no
one does!

Okay, if you're still here, then you're one of the few that do. This tool is
designed for you. (Well really, it's designed for me, but hopefully it'll help
you out.)

---
`imagend` is a simple and perscriptive tool for generating -- and re-generating
-- a large number of docker images. It uses a basic yaml definition, in
conjunction with [Golang templates](https://golang.org/pkg/text/template/) to
generate Dockerfiles, build and verify images, and push them to [https://hub.docker.com].

> **Perscriptive point:**
> Currently verification is based on the assumption that executing
> `<image> --version` in the built image will return `/.*<version>.*/`.
>
> An option would be to add the following to your template.
> ```
> RUN \
>   echo "echo \"{{ .Version }}\" > /usr/bin/{{ .Name }}" && \
>   chmod 755 /usr/bin/{{ .Name }}
> ```
>
> However you can always pass the `--skip-verify` flag, to skip this step.
>


### The Manifest

The [manifest.yml](manifest.yml.sample) file is a collection of image
definitions which support the following fields...

key | description
---|---
`name` | Name your image or image set. **required**
`image` | Name your image, if missing, `name` will be used. _optional_
`vesrion` | The version tag of the image you're building. **required**
`source` | Passed in to the template for use in defining a source image. _optional_
`native` | Passed in to the template to indicate that the source image contains the version being built. _optional_
`priority` | Determines image level build priority.This will only order by name, as duplicate names with varying versions/aliases are build concurrently. _optional_ / _default: 0_
`aliases` | A list of other image names to use in addition to the primary name. If additional tags are set, the alias images will get them as well. _optional_
`tags` | A list of additional tags to be applied to this image, e.g. `latest`. _optional_


#### The Template

`imagend` uses Go's built in templating system to generate Dockefiles for use
in building images.

> **Perscriptive point;** Templates are required to follow the naming convention
> of `Dockerfile.<image>.tmpl`, otherwise an error will be raised.

You can pretty much do whatever you want inside a template. It will have access
to everything that's in the [manifest.yml](manifest.yml.sample) file.

Here's a basic working example -- assuming the `source` passed exists...
```
# WARNING
# -------
# This file is dynamically generated, do not edit!
# See https://github.com/jmervine/imagend for generator details.

# name: {{ .Name }}
# version: {{ .Version }}
FROM {{.Source}}

{{if not .Native}}
RUN \
  echo "echo \"{{ .Version }}\" > /usr/bin/{{ .Name }}" && \
  chmod 755 /usr/bin/{{ .Name }}
{{end}}

CMD /bin/bash
```


### Development
```
# w/ make
make
make install

# w/o make
go get -v github.com/kardianos/govendor
govendor vet +local
govendor test +local
govendor install +local
```
