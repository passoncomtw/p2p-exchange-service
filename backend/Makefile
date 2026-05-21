API_FILE   := p2p.api
CONFIG     := etc/config.yaml
SWAGGER_OUT := docs/swagger.json
BINARY     := bin/p2p

.PHONY: dev swagger build clean

dev:
	go run p2p.go -f $(CONFIG)

swagger:
	@mkdir -p docs
	goctl api plugin \
		-plugin goctl-swagger="swagger -filename $(SWAGGER_OUT)" \
		-api $(API_FILE) \
		-dir .
	@echo "Swagger spec generated: $(SWAGGER_OUT)"

build:
	@mkdir -p bin
	go build -o $(BINARY) p2p.go
	@echo "Binary: $(BINARY)"

clean:
	rm -rf bin docs/swagger.json
