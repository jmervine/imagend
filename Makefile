GOVENDOR=$(shell echo "$(GOBIN)/govendor")

test: $(GOVENDOR) vet
	govendor test -v +local

vet: $(GOVENDOR)
	govendor vet +local

install:
	govendor install +local

build:
	govendor build -o ./bin/imagend -a +local

$(GOVENDOR):
	go get -v github.com/kardianos/govendor

samples:
	imagend -m manifest.yml.sample -o _samples -r
