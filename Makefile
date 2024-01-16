all: build/share/man/man1/dbkp.1

build:
	@mkdir -p build

build/dbkp build/docgen: build $(shell find . -type f -name "*.go")
	go build -o build ./...

build/share/man/man1/dbkp.1: build/docgen
	@build/docgen

clean:
	@rm -rf build
