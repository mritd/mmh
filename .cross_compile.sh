#!/usr/bin/env bash

set -e

DIST_PREFIX="mmh"
TARGET_DIR="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm linux/arm64"

rm -rf ${TARGET_DIR}
mkdir ${TARGET_DIR}

for pl in ${PLATFORMS}; do
    export GOOS=$(echo ${pl} | cut -d'/' -f1)
    export GOARCH=$(echo ${pl} | cut -d'/' -f2)
    export CGO_ENABLED=1
    export TARGET=${TARGET_DIR}/${DIST_PREFIX}_${GOOS}_${GOARCH}
    if [ "${GOOS}" == "windows" ]; then
        export TARGET=${TARGET_DIR}/${DIST_PREFIX}_${GOOS}_${GOARCH}.exe
    fi

    echo "build => ${TARGET}"
    go build -trimpath -o ${TARGET} \
            -ldflags    "-X 'github.com/mritd/mmh/pkg/cmd.Version=${BUILD_VERSION}' \
                        -X 'github.com/mritd/mmh/pkg/cmd.BuildDate=${BUILD_DATE}' \
                        -X 'github.com/mritd/mmh/pkg/cmd.CommitID=${COMMIT_SHA1}'\
                        -w -s"
done

