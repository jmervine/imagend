GOVENDOR=$(shell echo "$(GOBIN)/govendor")

vet: $(GOVENDOR)
	govendor vet -v +local

install: vet
	govendor install +local

$(GOVENDOR):
	go get -v github.com/kardianos/govendor
