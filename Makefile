build: test	build-no-test

build-no-test:
	go build -ldflags "-X github.com/swaros/contxt/module/configure.minversion=2 -X github.com/swaros/contxt/module/configure.midversion=5 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20221213.125547" -o ./bin/contxt cmd/cmd-contxt/main.go

install-local: build
	./bin/contxt run install-local

install-local-no-test: build-no-test
	./bin/contxt run install-local

clean:
	rm -f ./bin/contxt
	rm -rf ./dist

test:
	go test  -failfast ./module/configure/./...
	go test  -failfast ./module/dirhandle/./...
	go test  -failfast ./module/systools/./...
	go test  -failfast ./module/trigger/./...
	go test  -failfast ./module/taskrun/./...
	go test  -failfast ./module/awaitgroup/./...
	go test  -failfast ./module/shellcmd/./...

info: build
	./bin/contxt dir