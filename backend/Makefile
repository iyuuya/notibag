.PHONY: build clean

build:
	go build -o notibag-send ./cmd/send.go

clean:
	rm -f notibag-send

install: build
	mkdir -p ~/.local/bin && cp ./notibag-send ~/.local/bin
