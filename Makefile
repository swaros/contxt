build: test	build-no-test

	
build-no-test:
	go build -ldflags "-X github.com/swaros/contxt/configure.minversion=`cat docs/minversion` -X github.com/swaros/contxt/configure.midversion=`cat docs/midversion` -X github.com/swaros/contxt/configure.mainversion=`cat docs/mainversion` -X github.com/swaros/contxt/configure.build=`date -u +.%Y%m%d.%H%M%S`" -o ./bin/contxt cmd/cmd-contxt/main.go

install-local: build
	./bin/contxt run install-local

install-local-no-test: build-no-test
	./bin/contxt run install-local

clean:
	rm -f ./bin/contxt
	rm -rf ./dist

test-all:
	go test ./...

test:
	go test -timeout 30s github.com/swaros/contxt/taskrun
	go test -timeout 30s github.com/swaros/contxt/systools

info: build
	./bin/contxt dir
