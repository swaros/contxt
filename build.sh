echo "build script"

### build release

# compose build ldflags
BUILD_ARGS=" -X github.com/swaros/contxt/module/configure.minversion=0 -X github.com/swaros/contxt/module/configure.midversion=6 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20241121.130412-linux-release -X github.com/swaros/contxt/module/configure.shortcut=ctx -X github.com/swaros/contxt/module/configure.binaryName=contxt -X github.com/swaros/contxt/module/configure.cnShortCut=cn"
# build release
echo "-> build release"
go build -ldflags "$BUILD_ARGS" -o ./bin/contxt cmd/v2/main.go
./bin/contxt version
echo "-> build release done"