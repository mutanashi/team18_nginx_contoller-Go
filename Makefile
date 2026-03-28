BINARY=team18
INSTALL_PATH=/usr/local/bin

build:
	go build -o $(BINARY) .

install: build
	sudo mv $(BINARY) $(INSTALL_PATH)/

uninstall:
	sudo rm $(INSTALL_PATH)/$(BINARY)

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall clean
