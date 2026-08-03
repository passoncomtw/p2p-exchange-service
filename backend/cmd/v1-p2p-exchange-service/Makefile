API_FILE     := swagger.api
SWAGGER_DIST := internal/swagger/dist/spec.json
PATCH_SCRIPT := internal/swagger/patch-examples.py

.PHONY: swagger

swagger:
	@mkdir -p internal/swagger/dist
	goctl api plugin \
		-plugin goctl-swagger="swagger -filename $(SWAGGER_DIST)" \
		-api $(API_FILE) \
		-dir .
	@python3 $(PATCH_SCRIPT) $(SWAGGER_DIST)
	@echo "Swagger spec generated: $(SWAGGER_DIST)"
