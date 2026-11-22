SWAGGER_DIST := internal/swagger/dist

.PHONY: swagger

# 需先安裝 swag：go install github.com/swaggo/swag/cmd/swag@latest
swagger:
	swag init -g main.go -d ./,../../internal/response --parseInternal --outputTypes json -o $(SWAGGER_DIST)
	@echo "Swagger spec generated: $(SWAGGER_DIST)/swagger.json"
