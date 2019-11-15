BINARY			= utask

MAIN_LOCATION	= ./cmd

TEST_LOCATION	= ./...
TEST_CMD		= go test -v -mod=vendor -cover ${TEST_LOCATION}
TEST_CMD_COV	= ${TEST_CMD} -covermode=count -coverprofile=coverage.out

VERSION         = `git describe --tags $(git rev-list --tags --max-count=1)`
LAST_COMMIT		= `git rev-parse HEAD`
VERSION_PKG		= github.com/ovh/utask

DOCKER			= 0
DOCKER_OPT		=

define build_binary
	GO111MODULE=on go build -mod=vendor -ldflags "-X ${VERSION_PKG}.Commit=${LAST_COMMIT} -X ${VERSION_PKG}.Version=${VERSION}" \
		-o $(1) ${MAIN_LOCATION}/$(1)
	@[ ${DOCKER} -eq 0 ] || $(call docker_build,$(1))
endef

define docker_build
	docker build ${DOCKER_OPT} -f ${MAIN_LOCATION}/$(1)/Dockerfile .
endef

all: ${BINARY} 

${BINARY}: 
	$(call build_binary,${BINARY})

docker:
	@echo docker build enabled!
	$(eval DOCKER=1)

clean:
	rm -f ${BINARY}

re: clean all

release:
	bash hack/generate-install-script.sh

test:
	go get github.com/jstemmer/go-junit-report
	go get github.com/stretchr/testify/assert
	GO111MODULE=on DEV=true bash hack/test.sh ${TEST_CMD} 2>&1 | go-junit-report > report.xml

test-travis:
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	hack/test.sh ${TEST_CMD_COV}

test-docker: 
	DEV=true bash hack/test-docker.sh ${TEST_CMD}

run-test-stack:
	bash hack/test.sh bash hack/interactive.sh

run-test-stack-docker:
	bash hack/test-docker.sh bash hack/interactive.sh

package:
