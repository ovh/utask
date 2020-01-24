NG=ng
NPM=npm
NODE=node
TAR=tar -zcvf
UNTAR=tar -zxvf
ECHO=echo
DEL=rm -rf

DIST_DIR=dist
DIST_TAR=dist.tar.gz
DIR_NODEMODULES=node_modules

BASEHREF ?= /
PORT ?= 5557

all: debug clean install build

clean:
	$(DEL) $(DIR_NODEMODULES)
	$(DEL) $(DIST_DIR)
	$(DEL) $(DIST_TAR)

debug: 
	$(NPM) -v
	$(NODE) -v
	$(ECHO) "-=DEBUG=-"

install:
	$(NPM) install

dev:
	$(NODE) scripts/prebuild-env-dev.js
	$(NG) serve --open --port $(PORT) --base-href $(BASEHREF) --verbose=true

build: install
	$(NODE) scripts/prebuild-env-dev.js
	$(NG) build --base-href $(BASEHREF) --verbose=true
	# $(TAR) $(DIST_TAR) $(DIST_DIR)

build-prod: install
	$(NODE) scripts/prebuild-env-prod.js
	$(NG) build --prod --base-href $(BASEHREF) --verbose=true
	# $(TAR) $(DIST_TAR) $(DIST_DIR)

prod:
	$(NODE) scripts/prebuild-env-prod.js
	$(NG) serve --open --port $(PORT) --base-href $(BASEHREF) --prod

package:
	$(UNTAR) $(DIST_TAR)

test:
	$(NG) test
