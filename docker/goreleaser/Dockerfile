FROM scratch as binBase
COPY flytrap /

FROM alpine:edge
LABEL maintainer Urjit Singh Bhatia<(urjitsinghbhatia@gmail.com> (github: @urjitbhatia)

ARG VERSION=0.0.0

WORKDIR /usr/local/bin/

RUN     apk --no-cache add \
    ca-certificates

RUN buildDeps=' \
    bash \
    ' \
    set -x \
    && apk --no-cache add $buildDeps \
    && apk del $buildDeps \
    && echo "Build complete."

EXPOSE 9000
COPY --from=binBase /flytrap /usr/local/bin/flytrap
ENTRYPOINT [ "flytrap" ]
