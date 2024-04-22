BINARY_NAME=fewsatscli
RELEASE_DIR=$(BINARY_NAME)-v$(VERSION)
LDFLAGS=-ldflags "-X github.com/fewsats/fewsatscli/version.Commit=$(shell git describe --tags --always)"

PLATFORMS=darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm linux/arm64 windows/386 windows/amd64

.PHONY: all clean release

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/cli

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=X.Y.Z"; \
		exit 1; \
	fi
	@echo "Building release for version $(VERSION)"
	@$(MAKE) clean
	@$(MAKE) $(PLATFORMS)
	@echo "Compressing binaries"
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		binary_name=$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then \
			binary_name=$${binary_name}.exe; \
		fi; \
		tar_name=$(BINARY_NAME)-$$os-$$arch-v$(VERSION).tar.gz; \
		tar -czf $(RELEASE_DIR)/$$tar_name -C $(RELEASE_DIR) $$binary_name; \
		rm $(RELEASE_DIR)/$$binary_name; \
	done

$(PLATFORMS):
	@echo "Building for $@"
	@mkdir -p $(RELEASE_DIR)
	GOOS=$(firstword $(subst /, ,$@)) GOARCH=$(lastword $(subst /, ,$@)) go build -o $(RELEASE_DIR)/$(BINARY_NAME)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@)) ./cmd/cli
	@if [ "$(firstword $(subst /, ,$@))" = "windows" ]; then \
		mv $(RELEASE_DIR)/$(BINARY_NAME)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@)) $(RELEASE_DIR)/$(BINARY_NAME)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@)).exe; \
	fi

clean:
	@echo "Cleaning release directory"
	@rm -rf $(RELEASE_DIR)