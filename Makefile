# Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
# Note: redirected binaries to sasuser b/c homedir is full
EXECUTABLE=viya4-orders-cli
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64
MAIN=./main.go
VERSION=$(shell git describe --tags --abbrev=0 --always)
BUILDARGS=-s -w -X github.com/sassoftware/viya4-orders-cli/cmd.version=$(VERSION)
export GOFLAGS=-v -trimpath

win: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (Mac OS)

$(WINDOWS):
	@echo version: $(VERSION)
	env GOOS=windows GOARCH=amd64 go build -o /u/sapage/sasuser/dev/$(WINDOWS) -ldflags="$(BUILDARGS)" $(MAIN)
$(LINUX):
	@echo version: $(VERSION)
	env GOOS=linux GOARCH=amd64 go build -o /u/sapage/sasuser/dev/$(LINUX) -ldflags="$(BUILDARGS)" $(MAIN)
$(DARWIN):
	@echo version: $(VERSION)
	env GOOS=darwin GOARCH=amd64 go build -o /u/sapage/sasuser/dev/$(DARWIN) -ldflags="$(BUILDARGS)" $(MAIN)
clean:
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)
build: win linux darwin