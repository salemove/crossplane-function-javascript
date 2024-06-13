PROJECT_NAME = function-javascript
OUTPUT_DIR ?= bin

build: $(OUTPUT_DIR) generate
	@go build -o bin/function .

img.build: generate
	@docker build . --tag $(PROJECT_NAME)

xpkg.build: img.build
	@crossplane xpkg build -f package --embed-runtime-image=$(PROJECT_NAME)

generate:
	@go generate ./...

run: generate
	@go run . --insecure --debug

test: generate
	@go test -v ./...

vet: generate
	go vet

clean:
	@rm -rf $(OUTPUT_DIR)
	@find package -type f -iname '*.xpkg' -delete

$(OUTPUT_DIR):
	@mkdir -p "$(OUTPUT_DIR)"
