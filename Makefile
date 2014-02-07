DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps build

build:
	@mkdir -p bin/
	go get ./...
	gox -osarch="darwin/amd64" -osarch="linux/amd64" -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}" ../mirrorbox/...

deps:
	go get -d -v ./...
	echo $(DEPS) | xargs -n1 go get -d

update:
	go get -u -v
	echo $(DEPS) | xargs -n1 go get -d -u 

.PNONY: all deps 