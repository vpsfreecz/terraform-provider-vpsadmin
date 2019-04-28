TARGET = terraform-provider-vpsadmin

build:
	go build -o $(TARGET)

install:
	mkdir -p ~/.terraform.d/plugins
	cp -p $(TARGET) ~/.terraform.d/plugins/

.PHONY: build install
