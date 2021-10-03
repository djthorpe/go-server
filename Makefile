# Paths to packages
GO=$(shell which go)
SED=$(shell which sed)

# Paths to locations, etc
BUILD_DIR = "build"
PLUGIN_DIR = $(wildcard plugin/*)
BUILD_MODULE = "github.com/mutablelogic/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
BUILD_VERSION = $(shell git describe --tags)
BUILD_ARCH = $(shell $(GO) env GOARCH)
BUILD_PLATFORM = $(shell $(GO) env GOOS)

all: clean server plugins

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} ./cmd/server

plugins: $(PLUGIN_DIR)
#	@echo Build plugin media 
#	@${GO} get github.com/djthorpe/go-media
#	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/media.plugin ${BUILD_FLAGS} github.com/djthorpe/go-media/plugin/media

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

deb: nfpm go-server-httpserver-deb go-server-mdns-deb go-server-ldapauth-deb go-server-template-deb go-server-ddregister-deb

go-server-httpserver-deb: server plugin/httpserver plugin/log plugin/basicauth plugin/static plugin/env
	@echo Package go-server-httpserver deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-httpserver/nfpm.yaml > $(BUILD_DIR)/go-server-httpserver-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-httpserver-nfpm.yaml --packager deb --target $(BUILD_DIR)

go-server-mdns-deb: plugin/mdns
	@echo Package go-server-mdns deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-mdns/nfpm.yaml > $(BUILD_DIR)/go-server-mdns-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-mdns-nfpm.yaml --packager deb --target $(BUILD_DIR)

go-server-ldapauth-deb: plugin/ldapauth
	@echo Package go-server-ldapauth deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-ldapauth/nfpm.yaml > $(BUILD_DIR)/go-server-ldapauth-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-ldapauth-nfpm.yaml --packager deb --target $(BUILD_DIR)

go-server-template-deb: plugin/template plugin/text-renderer plugin/dir-renderer plugin/markdown-renderer
	@echo Package go-server-template deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-template/nfpm.yaml > $(BUILD_DIR)/go-server-template-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-template-nfpm.yaml --packager deb --target $(BUILD_DIR)

go-server-ddregister-deb: plugin/ddregister
	@echo Package go-server-ddregister deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-ddregister/nfpm.yaml > $(BUILD_DIR)/go-server-ddregister-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-ddregister-nfpm.yaml --packager deb --target $(BUILD_DIR)

nfpm:
	@echo Installing nfpm
	@${GO} mod tidy
	@${GO} install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.3.1	

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif
ifeq (,${SED})
        $(error "Missing sed binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
