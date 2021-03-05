build: test	
	go build -ldflags "-X github.com/swaros/contxt/context/configure.minversion=`cat docs/minversion` -X github.com/swaros/contxt/context/configure.midversion=`cat docs/midversion` -X github.com/swaros/contxt/context/configure.mainversion=`cat docs/mainversion` -X github.com/swaros/contxt/context/configure.build=`date -u +.%Y%m%d.%H%M%S`" -i -o ./bin/contxt cmd/cmd-contxt/main.go
	
	

install-local: test build
	cp ./bin/contxt ~/.local/bin/
	source <(contxt completion bash)

clean:
	rm -f ./bin/contxt
	rm -rf ./dist

test-all:
	go test -v ./...

test:
	go test -timeout 30s github.com/swaros/contxt/context/cmdhandle
	go test -timeout 30s github.com/swaros/contxt/context/systools
	go test -timeout 30s github.com/swaros/contxt/context/output

info: build
	./bin/contxt dir -info
