#!/usr/bin/env bash

set -e

TARGET_DIR="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm linux/arm64"
COMMANDS="mcp mcs mcx mec mgo mping mtun"

rm -rf ${TARGET_DIR}
mkdir ${TARGET_DIR}

for cmd in ${COMMANDS}; do
    if [ "$1" == "install" ]; then
        echo "install => ${GOPATH}/bin/${cmd}"
        go build -o ${GOPATH}/bin/${cmd} -ldflags \
            "-X 'github.com/mritd/mmh/cmd.Version=${BUILD_VERSION}' \
            -X 'github.com/mritd/mmh/cmd.BuildDate=${BUILD_DATE}' \
            -X 'github.com/mritd/mmh/cmd.CommitID=${COMMIT_SHA1}' \
            -X 'github.com/mritd/mmh/cmd.BuildCmd=${cmd}'"
    else
        for pl in ${PLATFORMS}; do
            export GOOS=$(echo ${pl} | cut -d'/' -f1)
            export GOARCH=$(echo ${pl} | cut -d'/' -f2)
            export CGO_ENABLED=1

            export TARGET=${TARGET_DIR}/${cmd}_${GOOS}_${GOARCH}
            if [ "${GOOS}" == "windows" ]; then
                export TARGET=${TARGET_DIR}/${cmd}_${GOOS}_${GOARCH}.exe
            fi

            echo "build => ${TARGET}"
            go build -trimpath -o ${TARGET} \
                    -ldflags    "-X 'github.com/mritd/mmh/cmd.Version=${BUILD_VERSION}' \
                                -X 'github.com/mritd/mmh/cmd.BuildDate=${BUILD_DATE}' \
                                -X 'github.com/mritd/mmh/cmd.CommitID=${COMMIT_SHA1}' \
                                -X 'github.com/mritd/mmh/cmd.BuildCmd=${cmd}' \
                                -w -s"
        done
    fi
done

