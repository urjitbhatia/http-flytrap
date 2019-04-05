FROM scratch as binBase
COPY flytrap /
COPY static /static
COPY templates /templates

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
COPY --from=binBase /templates /usr/local/bin/templates
COPY --from=binBase /static /usr/local/bin/static

# Tweak the handler TTL - how long data is retained for. (default 30mins)
# ENV HANDLER_TTL="30m"
ENTRYPOINT [ "flytrap" ]
