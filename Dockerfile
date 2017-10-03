FROM alpine:3.6
LABEL maintainer=devops@skuid.com
EXPOSE 3000
RUN apk add -U ca-certificates
ADD changelog /usr/local/bin/
ENTRYPOINT [ "changelog" ]
CMD ["serve"]