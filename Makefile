.PHONY: generate
generate:
	go generate -x ./generate/generator/main.go

.PHONY: build
build: generate
	go build -o rhino .

.PHONY: clean
clean:
	rm rhino