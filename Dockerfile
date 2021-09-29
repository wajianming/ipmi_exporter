ARG ARCH="amd64"
ARG OS="linux"
#FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc
FROM ubuntu:16.04
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY ./ipmi_exporter /bin/ipmi_exporter

RUN apt-get update -y \
&&  apt-get install freeipmi-tools -y

EXPOSE      9290
USER        nobody
ENTRYPOINT  [ "/bin/ipmi_exporter" ]
