GOVENDOR=$(shell echo "$(GOBIN)/govendor")

test: $(GOVENDOR) vet
	govendor test -v +local

vet: $(GOVENDOR)
	govendor vet +local

install: test
	govendor install +local

$(GOVENDOR):
	go get -v github.com/kardianos/govendor
