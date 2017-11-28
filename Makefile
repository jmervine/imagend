GOVENDOR=$(shell echo "$(GOBIN)/govendor")
GOFILES=$(shell find . -maxdepth 1 -type f -name "*.go" -not -name "*_test.go")

test: $(GOVENDOR) vet
	govendor test -v +local

vet: $(GOVENDOR)
	govendor vet +local

install: $(GOVENDOR)
	govendor install +local

build: $(GOVENDOR)
	govendor build -o ./bin/imagend -a +local

$(GOVENDOR):
	go get -v github.com/kardianos/govendor

samples:
	go run $(GOFILES) --manifest manifest.yml.sample --output _samples --remove
	go run $(GOFILES) --manifest manifest.yml.sample --docs
