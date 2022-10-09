# syntax=docker.io/docker/dockerfile:1

FROM gcr.io/distroless/base-debian11:latest

ARG BIN_DIR
ARG TARGETARCH
COPY $BIN_DIR/$TARGETARCH/linux_dvb_exporter /usr/bin/linux_dvb_exporter

EXPOSE 9111
ENTRYPOINT ["/usr/bin/linux_dvb_exporter"]
