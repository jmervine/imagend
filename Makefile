GOVENDOR=$(shell echo "$(GOBIN)/govendor")

vet: $(GOVENDOR)
	govendor vet +local

$(GOVENDOR):
	go get -v github.com/kardianos/govendor
