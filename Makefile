TARGET = terraform-provider-vpsadmin
PLUGIN_DIR = ~/.terraform.d/plugins
VERSION = 1.2.0
PLATFORM = linux_amd64
PROVIDER_DIR = terraform.vpsfree.cz/vpsfreecz/vpsadmin/$(VERSION)/$(PLATFORM)

build:
	go build -o $(TARGET)

docs:
	go generate

fmt:
	go fmt
	cd vpsadmin && go fmt

install:
	mkdir -p $(PLUGIN_DIR)/$(PROVIDER_DIR)
	cp -p $(TARGET) $(PLUGIN_DIR)/$(PROVIDER_DIR)/

.PHONY: build docs fmt install
