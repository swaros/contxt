build: test	build-no-test

build-no-test:
	go build -ldflags " -X github.com/swaros/contxt/module/configure.minversion=3 -X github.com/swaros/contxt/module/configure.midversion=5 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20231019.145117-linux-release" -o ./bin/contxt cmd/cmd-contxt/main.go
build-development:
	go build -ldflags " -X github.com/swaros/contxt/module/configure.minversion=3 -X github.com/swaros/contxt/module/configure.midversion=5 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20231019.145117-linux-release -X github.com/swaros/contxt/module/configure.shortcut=v2ctx -X github.com/swaros/contxt/module/configure.binaryName=ctxv2 -X github.com/swaros/contxt/module/configure.cnShortCut=v2cn" -o ./bin/ctxv2 cmd/v2/main.go
build-release:
	go build -ldflags " -X github.com/swaros/contxt/module/configure.minversion=3 -X github.com/swaros/contxt/module/configure.midversion=5 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20231019.145117-linux-release" -o ./bin/contxt cmd/cmd-contxt/main.go

clean:
	rm -f ./bin/contxt
	rm -rf ./dist

test:
	go test  -failfast ./module/yacl/./...
	go test  -failfast ./module/yamc/./...
	go test  -failfast ./module/runner/./...
	go test  -failfast ./module/ctxtcell/./...
	go test  -failfast ./module/configure/./...
	go test  -failfast ./module/dirhandle/./...
	go test  -failfast ./module/systools/./...
	go test  -failfast ./module/trigger/./...
	go test  -failfast ./module/linehack/./...
	go test  -failfast ./module/ctemplate/./...
	go test  -failfast ./module/ctxout/./...
	go test  -failfast ./module/taskrun/./...
	go test  -failfast ./module/awaitgroup/./...
	go test  -failfast ./module/shellcmd/./...
	go test  -failfast ./module/ctxshell/./...
	go test  -failfast ./module/tasks/./...
	go test  -failfast ./module/yaclint/./...
	go test  -failfast ./module/mimiclog/./...

info: build
	./bin/contxt dir