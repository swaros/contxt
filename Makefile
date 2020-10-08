build:
	go build -i -o ./bin/contxt cmd/cmd-contxt/main.go

install-local: build
	cp ./bin/contxt ~/.local/bin/

clean:
	rm -f ./bin/contxt