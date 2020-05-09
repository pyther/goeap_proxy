# This how we want to name the binary output
BINARY=goeap_proxy

# These are the values we want to pass for VERSION and BUILD
# git tag 1.0.1
# git commit -am "One more change after the tags"
VERSION=`git describe --tags --dirty --always`
BUILD=`date +%FT%T%z`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildStamp=${BUILD}"

# Builds the project
build:
	go build ${LDFLAGS} -o ${BINARY}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install
