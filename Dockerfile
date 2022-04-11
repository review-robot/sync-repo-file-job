FROM golang:latest as BUILDER

MAINTAINER xieweizhi<986740642@qq.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/sync-repo-file-job
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o sync-repo-file-job .

# copy binary config and utils
FROM alpine:3.14
COPY  --from=BUILDER /go/src/github.com/opensourceways/sync-repo-file-job/sync-repo-file-job /opt/app/sync-repo-file-job

ENTRYPOINT ["/opt/app/sync-repo-file-job"]