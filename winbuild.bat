echo "build script"
echo "-> build release"
go build -ldflags " -X github.com/swaros/contxt/module/configure.minversion=2 -X github.com/swaros/contxt/module/configure.midversion=6 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20241130.064109-linux-release -X github.com/swaros/contxt/module/configure.shortcut=ctx -X github.com/swaros/contxt/module/configure.binaryName=contxt -X github.com/swaros/contxt/module/configure.cnShortCut=cn" -o ./bin/contxt.exe cmd/v2/main.go
echo "-> build release done"