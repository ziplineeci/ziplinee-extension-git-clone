FROM alpine:3.13

LABEL maintainer="helm-ziplineeci.malsharbaji.com" \
        description="The ziplinee-extension-git-clone component is an Ziplinee extension to clone a git repository for builds handled by ZiplineeCI"

RUN apk add --update --no-cache \
      git \
      && rm -rf /var/cache/apk/*
COPY  publish/ziplinee-extension-git-clone /

ENV ZIPLINEE_LOG_FORMAT="console"

ENTRYPOINT ["/ziplinee-extension-git-clone"]