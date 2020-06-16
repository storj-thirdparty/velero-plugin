FROM golang:1.14-buster AS build
WORKDIR /go/src/github.com/jessicagreben/velero-plugin-for-storj-tardigrade
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/velero-plugin-for-storj-tardigrade .


FROM ubuntu:bionic
RUN mkdir /plugins
COPY --from=build /go/bin/velero-plugin-for-storj-tardigrade /plugins/
USER nobody:nogroup
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/* /target/."]
