OUTPUT = slack-status
GOFILES = *.go go.mod go.sum
DIST_ARCH = darwin-amd64 darwin-arm64 linux-amd64
DIST_TARGETS = $(addprefix dist/$(OUTPUT)-,$(DIST_ARCH))
VERSION ?= $(shell git describe --tags --dirty)

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

slack-status: $(GOFILES) certs
	go build -o $(OUTPUT)

.PHONY: dist
dist: $(DIST_TARGETS)

$(DIST_TARGETS): $(GOFILES) certs
	@mkdir -p ./dist
	GOOS=$(word 3, $(subst -, ,$(@))) GOARCH=$(word 4, $(subst -, ,$(@))) \
		 go build \
		 -ldflags '-X "main.version=${VERSION}" -X "main.defaultClientID=$(CLIENT_ID)" -X "main.defaultClientSecret=$(CLIENT_SECRET)"' \
		 -o $@

.PHONY: certs
certs: certs/key.pem
certs/cert.pem: certs/key.pem
certs/key.pem:
	mkdir -p certs/
	openssl req -x509 -subj "/C=US/O=Slack Status CLI/CN=localhost:8888" \
		-nodes -days 365 -newkey "rsa:2048" \
		-addext "subjectAltName=DNS:localhost:8888" \
		-keyout certs/key.pem -out certs/cert.pem

.PHONY: clean
clean:
	rm -f ./slack-status
	rm -fr ./dist
	rm -fr ./certs

.PHONY: install-hooks
install-hooks:
	pre-commit install --overwrite --install-hooks
