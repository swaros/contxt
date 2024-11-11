echo "build script"

### build release

# compose build ldflags
BUILD_ARGS=" -X github.com/swaros/contxt/module/configure.minversion=0 -X github.com/swaros/contxt/module/configure.midversion=6 -X github.com/swaros/contxt/module/configure.mainversion=0 -X github.com/swaros/contxt/module/configure.build=.20241111.061200-linux-release"
# build release
echo "-> build release"
go build -ldflags "$BUILD_ARGS" -o ./bin/contxt cmd/cmd-contxt/main.go
./bin/contxt version
echo "-> build release done"