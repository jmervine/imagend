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

#
# find a better way
run:
	go run $(shell find . -name "*.go" -not -name "*_test.go" | grep -v "vendor")

help:
	go run $(shell find . -name "*.go" -not -name "*_test.go" | grep -v "vendor") -h
