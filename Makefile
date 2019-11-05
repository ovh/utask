BINARY			= utask

MAIN_LOCATION	= ./cmd
TEST_LOCATION	= ./...

TEST_CMD		= go test -v -cover ${TEST_LOCATION}

VERSION			= 1.0.0
LAST_COMMIT		= `git rev-parse HEAD`
VERSION_PKG		= github.com/ovh/utask

DOCKER			= 0
DOCKER_OPT		=

define build_binary
	GO111MODULE=on go get -d -v  ./...
	GO111MODULE=on go build -ldflags "-X ${VERSION_PKG}.Commit=${LAST_COMMIT} -X ${VERSION_PKG}.Version=${VERSION}" \
		-o $(1) ${MAIN_LOCATION}/$(1)
	@[ ${DOCKER} -eq 0 ] || $(call docker_build,$(1))
endef

define docker_build
	docker build ${DOCKER_OPT} -f ${MAIN_LOCATION}/$(1)/Dockerfile .
endef

all: ${BINARY} 

${BINARY}: 
	$(call build_binary,${BINARY})

generate:
	go get github.com/ybriffa/jsonenums
	go get golang.org/x/tools/cmd/stringer
	go generate ./...

docker:
	@echo docker build enabled!
	$(eval DOCKER=1)

clean:
	rm -f ${BINARY}

re: clean all

test:
	go get github.com/jstemmer/go-junit-report
	go get github.com/stretchr/testify/assert
	GO111MODULE=on DEV=true bash test.sh ${TEST_CMD} 2>&1 | go-junit-report > report.xml

test-docker: 
	DEV=true bash test-docker.sh ${TEST_CMD} 

run-test-stack:
	bash test.sh bash interactive.sh

run-test-stack-docker:
	bash test-docker.sh bash interactive.sh

package: