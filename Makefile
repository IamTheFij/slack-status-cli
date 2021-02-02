OUTPUT = slack-status
GOFILES = *.go go.mod go.sum
DIST_ARCH = darwin-amd64 linux-amd64
DIST_TARGETS = $(addprefix dist/$(OUTPUT)-,$(DIST_ARCH))

.PHONY: default
default: slack-status

.PHONY: run
run: slack-status
	./slack-status

.PHONY: all
all: dist

.PHONY: test
test:
	pre-commit run --all-files
	go test

slack-status: $(GOFILES)
	go build -o $(OUTPUT)

.PHONY: dist
dist: $(DIST_TARGETS)

$(DIST_TARGETS): $(GOFILES)
	@mkdir -p ./dist
	GOOS=$(word 3, $(subst -, ,$(@))) GOARCH=$(word 4, $(subst -, ,$(@))) \
		 go build \
		 -ldflags '-X "main.defaultClientID=$(CLIENT_ID)" -X "main.defaultClientSecret=$(CLIENT_SECRET)"' \
		 -o $@

.PHONY: clean
clean:
	rm ./slack-status
	rm -fr ./dist

.PHONY: install-hooks
install-hooks:
	pre-commit install --overwrite --install-hooks
