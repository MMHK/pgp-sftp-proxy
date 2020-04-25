FROM golang:1.13-alpine as builder

# Add Maintainer Info
LABEL maintainer="Sam Zhou <sam@mixmedia.com>"

# Set the Current Working Directory inside the container
WORKDIR /app/pgp-sftp-proxy

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go version \
 && export GO111MODULE=on \
 && export GOPROXY=https://goproxy.io \
 && go mod vendor \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pgp-sftp-proxy

######## Start a new stage from scratch #######
FROM alpine:latest  

ENV HOST=0.0.0.0:3333 \
 ROOT=/root/pgp-sftp-proxy/web_root \
 TEMP=/tmp \
 SSH_HOST= \
 SSH_USER= \
 SSH_PWD= \
 SSH_KEY= \
 DEPLOY_PATH_DEV=/Interface_Development_Files/ \
 DEPLOY_PATH_PRODUCTION=/Interface_Production_Files/ \
 DEPLOY_PATH_TESTING=/Interface_UAT_Files/ 

RUN apk --no-cache add ca-certificates \
    && apk add --update python python-dev py-pip build-base \
    && pip install dumb-init \
    && apk del python  python-dev py-pip build-base \
    && rm -rf /var/cache/apk/* \
    && rm -rf /tmp/*

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/pgp-sftp-proxy .
 
EXPOSE 3333

ENTRYPOINT ["dumb-init"]

CMD if ! which envsubst > /dev/null 2>&1; then envsubst() { while read line; do line=$( echo $line | sed 's/"/\\"/g' ); eval echo $line; done; }; fi \
 && /usr/local/sbin/envsubst < /root/pgp-sftp-proxy/config.json > /root/pgp-sftp-proxy/temp.json \
 && /root/pgp-sftp-proxy/pgp-sftp-proxy -c /root/pgp-sftp-proxy/temp.json
