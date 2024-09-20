echo "build script"
echo "-> build release"
go build -ldflags " -X github.com/swaros/contxt/module/configure.minversion=3 -X github.com/swaros/contxt/module/configure.midversion=5 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20240920.060144-linux-release" -o ./bin/contxt.exe cmd/cmd-contxt/main.go
echo "-> build release done"