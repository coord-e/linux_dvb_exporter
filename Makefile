BIN := linux_dvb_exporter

.PHONY: build
build: $(BIN)

GO_FILES := $(shell find . -type f -name '*.go' -print)

ifdef RELEASE
	# XXX: In principle, the build does not require a call to the C compiler or a link to libc, since Cgo is only used to extract structures and constants.
	#      But Cgo is not designed to work without them, so linux_dvb_exporter gives up static linking.
	GO_LDFLAGS += -w -s
	GO_FLAGS += -a
endif

VERSION := $(shell cat VERSION)
COMMIT_SHA := $(shell git rev-parse --short HEAD)
GO_LDFLAGS += -X 'main.BuildVersion=$(VERSION)' -X 'main.BuildCommitSha=$(COMMIT_SHA)'

$(BIN): $(GO_FILES)
	CGO_ENABLED=1 go build -o $@ -tags=$(GO_BUILD_TAGS) $(GO_FLAGS) -ldflags "$(GO_LDFLAGS)"

.PHONY: clean
clean:
	$(RM) $(BIN)
