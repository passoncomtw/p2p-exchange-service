API_FILE     := swagger.api
SWAGGER_DIST := internal/swagger/dist/spec.json

.PHONY: swagger

swagger:
	@mkdir -p internal/swagger/dist
	goctl api plugin \
		-plugin goctl-swagger="swagger -filename $(SWAGGER_DIST)" \
		-api $(API_FILE) \
		-dir .
	@echo "Swagger spec generated: $(SWAGGER_DIST)"
