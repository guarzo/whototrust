# Makefile

RESOURCE_GROUP=flgyd
APP_NAME=whototrust
ENVIRONMENT_NAME=flygd-analytics
ACR_NAME=flygd
REGISTRY_NAME=flygd
REPOSITORY_NAME=flygd

.PHONY: all build acr_build update_aca full_deploy run add_secrets clean help

all: build

build:
	go build -o $(APP_NAME)

acr_build: increment_version
	@VERSION=$$(cat VERSION) && \
	az acr build --image $(REPOSITORY_NAME)/$(APP_NAME):$${VERSION} --registry $(REGISTRY_NAME) --build-arg VERSION=$${VERSION} .


update_aca:
	@VERSION=$$(cat VERSION) && \
	az containerapp update -n $(APP_NAME) -g $(RESOURCE_GROUP) --image $(REGISTRY_NAME).azurecr.io/$(REPOSITORY_NAME)/$(APP_NAME):$${VERSION}

full_deploy: acr_build update_aca

run:
	EVE_CLIENT_ID=${EVE_CLIENT_ID} EVE_CLIENT_SECRET=${EVE_CLIENT_SECRET} EVE_CALLBACK_URL=http://localhost:8080/callback go run main.go

add_secrets:
	@if [ -z "$(SECRET_NAME)" ] || [ -z "$(SECRET_VALUE)" ]; then \
		echo "SECRET_NAME and SECRET_VALUE must be provided"; \
		exit 1; \
	fi
	az containerapp secret set --name $(APP_NAME) --resource-group $(RESOURCE_GROUP) --secrets "$(SECRET_NAME)=$(SECRET_VALUE)"

clean:
	go clean
	rm -rf data
	rm -f $(APP_NAME)

help:
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build             Build the Go application"
	@echo "  acr_build         Build the container image using Azure Container Registry"
	@echo "  deploy            Deploy the application using the updated app.yaml"
	@echo "  full_deploy       Build the container image, update the app.yaml file, and deploy the application"
	@echo "  update_config     Update the container app configuration with app.yaml"
	@echo "  run               Run the Go application with environment variables"
	@echo "  add_secrets       Add secrets to the container app (requires SECRET_NAME and SECRET_VALUE)"
	@echo "  clean             Clean up generated files"
	@echo "  help              Show this help message"

increment_version:
	@VERSION=$$(cat VERSION) && \
	NEW_VERSION=$$(echo $${VERSION} | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}') && \
	echo $${NEW_VERSION} > VERSION && \
	echo "Updated version to $${NEW_VERSION}"
