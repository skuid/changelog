FROM alpine:3.9
LABEL maintainer=devops@skuid.com
# Update this git sha with 'git rev-parse head' on build
LABEL "org.label-schema.vcs-ref"=92af5750c7289607e9a74a324bda8e03d792a5c8
EXPOSE 3000
RUN apk add -U ca-certificates
ADD changelog /usr/local/bin/
ENTRYPOINT [ "changelog" ]
CMD ["serve"]