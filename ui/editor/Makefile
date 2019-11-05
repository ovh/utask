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
	$(NG) serve --open --port 5555 --base-href $(BASEHREF)

build: install
	$(NG) build --base-href $(BASEHREF)
	# $(TAR) $(DIST_TAR) $(DIST_DIR)

build-prod: install
	$(NG) build --prod --base-href $(BASEHREF)
	# $(TAR) $(DIST_TAR) $(DIST_DIR)

prod:
	$(NG) serve --open --port 5555 --prod --base-href $(BASEHREF)
	$(TAR) $(DIST_TAR) $(DIST_DIR)

package:
	$(UNTAR) $(DIST_TAR)

test:
	$(ECHO) "No Test for the moment"
