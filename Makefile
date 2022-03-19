BINARY ?= ./

.PHONY: binary
binary: $(BINARY)

$(BINARY):
	go build \
		-trimpath \
		-o $(BINARY) \
		main.go

.PHONY: binaries
binaries:
	$(MAKE) OS=linux ARCH=amd64 ARCHNAME=x86_64 versioned-binary
	$(MAKE) OS=linux ARCH=arm64 versioned-binary
	$(MAKE) OS=darwin ARCH=amd64 ARCHNAME=x86_64 versioned-binary
	$(MAKE) OS=darwin ARCH=arm64 versioned-binary

.PHONY: versioned-binary
versioned-binary:
	GOOS=$(OS) GOARCH=$(ARCH) $(MAKE) BINARY=stakeout-$(or ${ARCHNAME},${ARCHNAME},${ARCH})-$(OS) binary
