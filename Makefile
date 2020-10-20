build: test
	go build -i -o ./bin/contxt cmd/cmd-contxt/main.go

install-local: test build
	cp ./bin/contxt ~/.local/bin/

clean:
	rm -f ./bin/contxt

test-all:
	go test -v ./...

test:
	go test -timeout 30s github.com/swaros/contxt/context/cmdhandle -v
	go test -timeout 30s github.com/swaros/contxt/context/systools -v

info: build
	./bin/contxt dir -info
