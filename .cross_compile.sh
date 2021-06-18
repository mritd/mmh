#!/usr/bin/env bash

set -e

TARGET_DIR="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm linux/arm64"
COMMANDS="mcp mcs mcx mec mgo mping mtun"

rm -rf ${TARGET_DIR}
mkdir ${TARGET_DIR}

if [ "$1" == "install" ]; then
    echo "install => ${GOPATH}/bin/mmh"
    go build -o ${GOPATH}/bin/mmh -ldflags \
        "-X 'github.com/mritd/mmh/cmd.Version=${BUILD_VERSION}' \
        -X 'github.com/mritd/mmh/cmd.BuildDate=${BUILD_DATE}' \
        -X 'github.com/mritd/mmh/cmd.CommitID=${COMMIT_SHA1}'"
    for cmd in ${COMMANDS}; do
        echo "install => ${GOPATH}/bin/${cmd}"
        ln -sf ${GOPATH}/bin/mmh ${GOPATH}/bin/${cmd}
    done
elif [ "$1" == "uninstall" ]; then
    echo "remove => ${GOPATH}/bin/mmh"
    rm -f ${GOPATH}/bin/mmh
    for cmd in ${COMMANDS}; do
        echo "remove => ${GOPATH}/bin/${cmd}"
        rm -f ${GOPATH}/bin/${cmd}
    done
elif [ "$1" == "completion" ]; then
    for s in bash zsh fish powershell; do
        mmh --completion ${s} > docs/completion/mmh.${s}
    done
    cat docs/completion/mmh.zsh > docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mmh' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mcs' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mcx' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mcp' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mec' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mgo' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mping' >> docs/completion/mmh.ohmyzsh
    echo 'compdef _mmh mtun' >> docs/completion/mmh.ohmyzsh
else
    for pl in ${PLATFORMS}; do
        export GOOS=$(echo ${pl} | cut -d'/' -f1)
        export GOARCH=$(echo ${pl} | cut -d'/' -f2)
        export CGO_ENABLED=0

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

