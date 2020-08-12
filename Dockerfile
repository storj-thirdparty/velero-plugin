FROM golang:1.14-buster AS build
ENV GOPROXY=https://proxy.golang.org
WORKDIR /storj
COPY . .
RUN make go-build

FROM ubuntu:bionic
RUN mkdir /plugins
COPY --from=build /storj/velero-plugin-storj /plugins/
USER nobody:nogroup
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/* /target/."]
