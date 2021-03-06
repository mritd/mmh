BUILD_VERSION   := $(shell cat version)
BUILD_DATE      := $(shell date "+%F %T")
COMMIT_SHA1     := $(shell git rev-parse HEAD)

all: clean
	bash .cross_compile.sh
	tar -C docs -zcvf dist/addons.tar.gz addons
	tar -C docs -zcvf dist/completions.tar.gz completions

release: clean all
	ghr -u mritd -t $(GITHUB_TOKEN) -replace -recreate --debug ${BUILD_VERSION} dist

pre-release: clean all
	ghr -u mritd -t $(GITHUB_TOKEN) -replace -recreate -prerelease --debug ${BUILD_VERSION} dist

clean:
	rm -rf dist

install:
	bash .cross_compile.sh install

uninstall:
	bash .cross_compile.sh uninstall

completion:
	bash .cross_compile.sh completion

.PHONY: all release pre-release clean install uninstall completion
