all:
	@echo "nothing is done when 'all' is done, try make help"

help:	## show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

vPATH := $(shell echo ${PATH})
CERTS_PATH := $(shell if [ -z "${CERTS_PATH}" ] ; then echo /local/tmp/certs ; else echo ${CERTS_PATH} ; fi)
CERT_SERVER := $(shell if [ -z "${CERT_SERVER}" ] ; then hostname ; else echo ${CERT_SERVER} ; fi)
VERSION := $(shell git describe --tags --always)

.PHONY: test
test: export GO_TEST_LOG_LEVEL = ERROR
test:	## go test the application
	go test -v ./...

.PHONY: test-verbose
test-verbose:	## go test the application
	go test -v ./...

.PHONY: test-cleansftp
test-cleansftp:	## go test clean sftp
	go test -v -count=1 -run '^TestBasicSftpDryrunSynczer' ./internal/walker

.PHONY: test-this
test-this:	## go test the application
	go test -v -run TestMakeTestFilesTree github.com/t-beigbeder/vdasync/internal/common

.PHONY: build-this
build-this:	## go build only this
build-this: export GOOS = linux
build-this: export GOARCH = amd64
build-this: export BDIR = bin/lamd64
build-this: export BEXT =
build-this:
	go build -o $(BDIR)/vdasftp$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/plugins/sftp/main.go

.PHONY: test-again
test-again:	export OTVL_TEST_FULL = 1
test-again: export GO_TEST_LOG_LEVEL = ERROR
test-again:	## go test the application again
	go test -v -count=1 -run '^TestBasicSftpDryrunSynczer' ./internal/walker
	go test -v -count=1 ./...

.PHONY: test-again-verbose
test-again-verbose:	export OTVL_TEST_FULL = 1
test-again-verbose:	## go test the application again
	go test -v -count=1 ./...

.PHONY: xbuild
xbuild:
	go build -o $(BDIR)/vdasync$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/vdasync/main.go
	go build -o $(BDIR)/vdaservice$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/vdaservice/main.go
	go build -o $(BDIR)/testcerts$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/testcerts/main.go
	go build -o $(BDIR)/vdaserver$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/vdaserver/main.go
	go build -o $(BDIR)/vdatserver$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/vdatserver/main.go
	go build -o $(BDIR)/vdas3$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/plugins/s3/main.go
	go build -o $(BDIR)/vdaencrypt$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/plugins/encrypt/main.go
	go build -o $(BDIR)/vdasftp$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/plugins/sftp/main.go
	go build -o $(BDIR)/localFiles$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/plugins/localfiles/main.go
	go build -o $(BDIR)/vdasftpsync$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/vdasftpsync/main.go
	go build -o $(BDIR)/csavsftp$(BEXT) -ldflags "-X github.com/t-beigbeder/vdasync/config.Version=$(VERSION)" cmd/plugins/examples/csavsftp/main.go

.PHONY: build
build: export GOOS = linux
build: export GOARCH = amd64
build: export BDIR = bin/lamd64
build: export BEXT =
build: xbuild

.PHONY: wbuild
wbuild: export GOOS = windows
wbuild: export GOARCH = amd64
wbuild: export BDIR = bin/wamd64
wbuild: export BEXT = .exe
wbuild: xbuild

.PHONY: release
release:	## go build and release tgz
	make build
	make wbuild
	mkdir -p tmp
	tar czf tmp/vdasync-linux-amd64-$(VERSION).tgz --exclude .gitignore -C bin/lamd64 vdasync vdaservice vdaserver vdatserver
	tar czf tmp/vdasync-all-linux-amd64-$(VERSION).tgz --exclude .gitignore -C bin/lamd64 .
	tar czf tmp/vdasync-sftp-windows-amd64-$(VERSION).tgz --exclude .gitignore -C bin/wamd64 vdasftpsync.exe
	tar czf tmp/vdasync-all-windows-amd64-$(VERSION).tgz --exclude .gitignore -C bin/wamd64 .

.PHONY: certs
certs:	## generate test certificates
	@mkdir -p $(CERTS_PATH)
	go build -o bin/testcerts cmd/testcerts/main.go
	bin/testcerts -ca $(CERTS_PATH)/sca-cert.pem -cakey $(CERTS_PATH)/sca-key.pem -cn Server-CA
	@echo openssl x509 -in $(CERTS_PATH)/sca-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/cca-cert.pem -cakey $(CERTS_PATH)/cca-key.pem -cn Client-CA
	@echo openssl x509 -in $(CERTS_PATH)/cca-cert.pem -text -noout
	bin/testcerts -cert $(CERTS_PATH)/self-cert.pem -key $(CERTS_PATH)/self-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/self-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/sca-cert.pem -cakey $(CERTS_PATH)/sca-key.pem -hosts localhost,$(CERT_SERVER) -cert $(CERTS_PATH)/localhost-cert.pem -key $(CERTS_PATH)/localhost-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/localhost-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/cca-cert.pem -cakey $(CERTS_PATH)/cca-key.pem -hosts localhost -cert $(CERTS_PATH)/plugin-cert.pem -key $(CERTS_PATH)/plugin-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/plugin-cert.pem -text -noout
	bin/testcerts -ca $(CERTS_PATH)/cca-cert.pem -cakey $(CERTS_PATH)/cca-key.pem -cert $(CERTS_PATH)/client-cert.pem -key $(CERTS_PATH)/client-key.pem
	@echo openssl x509 -in $(CERTS_PATH)/client-cert.pem -text -noout

.PHONY: format
format:	## format go code
	gofmt -w .

.PHONY: grpc-code
grpc-code: export PATH=$(vPATH):$(HOME)/.local/bin:/home/dv-user/go/bin
grpc-code:	## generate grpc code from proto files
	protoc --go_out=. --go-grpc_out=. grpc/ope.proto
	protoc --go_out=. --go-grpc_out=. grpc/dssa.proto
	protoc --go_out=. --go-grpc_out=. grpc/opelog.proto