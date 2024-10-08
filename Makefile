VERSION=`git describe --tags --dirty="-dev" | sed 's/[0-9]*-g//'`
BUILD=`date +%FT%T%z`

LDFLAGS=-X github.com/anomalous69/FChannel/config.Version=${VERSION} -X github.com/anomalous69/FChannel/config.BuildTime=${BUILD}
FLAGS=-ldflags "-w -s ${LDFLAGS}"
FLAGS_DEBUG=-ldflags "${LDFLAGS}"

debug:
	go build -o fchan ${FLAGS_DEBUG}

build:
	go build -o fchan ${FLAGS}

clean:
	if [ -f "fchan" ]; then rm "fchan"; fi

.PHONY: clean install
