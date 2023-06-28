BINARY			= utask

MAIN_LOCATION	= ./cmd

TEST_LOCATION	?= ./...
# timeout for go test is here to prevent tests running infinitely if one runResolution leads to a step that never recovers (e.g. missing push on a stepChan)
# 30 seconds per unit tests should be enough
TEST_CMD		= go test -count=1 -timeout 30s -v -cover -p 1 ${TEST_LOCATION}
TEST_CMD_COV	= ${TEST_CMD} -covermode=count -coverprofile=coverage.out

SOURCE_FILES 	= $(shell find ./ -type f -name "*.go" | grep -v _test.go)

VERSION 		:= $(shell git describe --exact-match --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2> /dev/null)
ifndef VERSION
	VERSION = $(shell git describe --abbrev=3 --tags $(git rev-list --tags --max-count=1))-dev
endif

LAST_COMMIT		:= $(shell git rev-parse HEAD)
VERSION_PKG		= github.com/ovh/utask

DOCKER			= 0
DOCKER_OPT		=

define goreleaser
	VERSION_PKG=${VERSION_PKG} LASTCOMMIT=${LAST_COMMIT} VERSION=${VERSION} goreleaser --rm-dist $(1)
endef

define build_binary
	GO111MODULE=on go build -ldflags "-X ${VERSION_PKG}.Commit=${LAST_COMMIT} -X ${VERSION_PKG}.Version=${VERSION}" \
		-o $(1) ${MAIN_LOCATION}/$(1)
	@[ ${DOCKER} -eq 0 ] || $(call docker_build,$(1))
endef

define docker_build
	docker build ${DOCKER_OPT} -f ${MAIN_LOCATION}/$(1)/Dockerfile .
endef

all: ${BINARY}

${BINARY}: $(SOURCE_FILES) go.mod
	$(call build_binary,${BINARY})

docker:
	@echo docker build enabled!
	$(eval DOCKER=1)

clean:
	rm -f ${BINARY}

re: clean all

release:
	bash hack/generate-install-script.sh

release-utask-lib:
	cd ui/dashboard/projects/utask-lib && npm version $(VERSION) --allow-same-version
	cd ui/dashboard && npm ci && ng build --prod utask-lib
	npm publish ui/dashboard/dist/utask-lib --access public

test:
	# moving to another location to go get some packages, otherwise it will include those packages as dependencies in go.mod
	cd ${HOME} && go install github.com/jstemmer/go-junit-report/v2@latest
	GO111MODULE=on DEV=true bash hack/test.sh ${TEST_CMD} 2>&1 | go-junit-report | tee report.xml

test-dev:
	GO111MODULE=on DEV=true bash hack/test.sh ${TEST_CMD} 2>&1

test-travis:
	# moving to another location to go get some packages, otherwise it will include those packages as dependencies in go.mod
	cd ${HOME} && go go install github.com/mattn/goveralls@latest
	hack/test.sh ${TEST_CMD_COV}

test-docker:
	cd ${HOME} && go install github.com/mattn/goveralls@latest
	DEV=true bash hack/test-docker.sh ${TEST_CMD}

run-test-stack:
	bash hack/test.sh bash hack/interactive.sh

run-test-stack-docker:
	bash hack/test-docker.sh bash hack/interactive.sh

run-goreleaser:
	export BINDIR=${GOPATH}/bin; go install github.com/goreleaser/goreleaser@v1.6.3
	rm -rf .cache
ifneq (,$(findstring -dev,$(VERSION)))
	@echo Run Goreleaser in snapshot mod
	$(call goreleaser,--snapshot)
else
	@echo Run Goreleaser in release mod
	$(call goreleaser)
endif

package:

makefile:
	sed -e 's/VERSION=/VERSION=${VERSION}/g' hack/Makefile-child | sed -e 's/LAST_COMMIT=/LAST_COMMIT=${LAST_COMMIT}/g' >| Makefile  

.PHONY: all clean test re package release test test-travis test-docker run-test-stack run-test-stack-docker run-goreleaser docker makefile
