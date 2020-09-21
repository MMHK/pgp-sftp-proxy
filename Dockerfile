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
 ROOT=/app/web_root \
 TEMP=/tmp \
 SSH_HOST= \
 SSH_USER= \
 SSH_PWD= \
 SSH_KEY= \
 DEPLOY_PATH_DEV=/Interface_Development_Files/ \
 DEPLOY_PATH_PRODUCTION=/Interface_Production_Files/ \
 DEPLOY_PATH_TESTING=/Interface_UAT_Files/ 

RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
 && chmod +x /usr/local/bin/dumb-init \
 && apk add --update libintl \
 && apk add --virtual build_deps gettext \
 && cp /usr/bin/envsubst /usr/local/bin/envsubst \
 && apk del build_deps

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/pgp-sftp-proxy/pgp-sftp-proxy .
COPY --from=builder /app/pgp-sftp-proxy/web_root ./web_root
COPY --from=builder /app/pgp-sftp-proxy/config.json .
 
EXPOSE 3333

ENTRYPOINT ["dumb-init"]

CMD envsubst < /app/config.json > /app/temp.json \
 && /app/pgp-sftp-proxy -c /app/temp.json
