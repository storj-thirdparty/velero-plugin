FROM golang:1.14-buster AS build
ENV GOPROXY=https://proxy.golang.org
COPY . /go/src/velero-plugin-for-tardigrade
WORKDIR /go/src/velero-plugin-for-tardigrade
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/velero-plugin-for-tardigrade ./cmd

FROM ubuntu:bionic
RUN mkdir /plugins
COPY --from=build /go/bin/velero-plugin-for-tardigrade /plugins/
USER nobody:nogroup
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/* /target/."]
