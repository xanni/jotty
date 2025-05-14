sources = keybindings.ditaa permascroll.ebnf
targets = $(addsuffix .png, $(basename $(sources)))

all: jotty $(targets)

clean:
	-rm jotty $(targets)

test:
	go test ./...

.PHONY: all clean test

version.txt:
	go generate ./...

jotty: jotty.go */*.go go.mod version.txt
	go build -ldflags="-s -w"
	-upx --lzma jotty

$(targets): $(sources)
	plantuml $?
	-optipng $(targets)
