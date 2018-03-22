FROM debian:jessie

ENV HOST=0.0.0.0:3333 \
 ROOT=/usr/local/pgp-sftp-proxy/web_root \
 TEMP=/tmp \
 SSH_HOST= \
 SSH_USER= \
 SSH_PWD= \
 SSH_KEY= \
 DEPLOY_PATH_DEV= \
 DEPLOY_PATH_PRODUCTION= \
 DEPLOY_PATH_TESTING=
 
WORKDIR /root/src/github.com/mmhk/pgp-sftp-proxy

COPY . .

RUN set -x  \
# Install runtime dependencies
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        gettext-base \
# install go runtime
 && curl -O https://dl.google.com/go/go1.9.4.linux-amd64.tar.gz \
 && tar xvf go1.9.4.linux-amd64.tar.gz \
 && mv ./go /usr/local/go \
# build pgp-sftp-proxy
 && export GOPATH=/root \
 && export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin \
 && go get -v \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pgp-sftp-proxy . \
 && mkdir /usr/local/pgp-sftp-proxy \
 && mv web_root /usr/local/pgp-sftp-proxy/web_root \
 && mv config.json /usr/local/pgp-sftp-proxy/config.json \
 && mv pgp-sftp-proxy /usr/bin/pgp-sftp-proxy \
# Install dumb-init (to handle PID 1 correctly).
# https://github.com/Yelp/dumb-init
 && curl -Lo /tmp/dumb-init.deb https://github.com/Yelp/dumb-init/releases/download/v1.1.3/dumb-init_1.1.3_amd64.deb \
 && dpkg -i /tmp/dumb-init.deb \
# Clean up
 && apt-get purge --auto-remove -y \
        curl git \
 && apt-get clean \
 && rm -rf /tmp/* /var/lib/apt/lists/* \
 && rm -Rf /root/src \
 && rm -Rf /root/bin \
 && rm -Rf /root/pkg \
 && rm -Rf /usr/local/go 
 
EXPOSE 3333

ENTRYPOINT ["dumb-init"]

CMD envsubst < /usr/local/pgp-sftp-proxy/config.json > /usr/local/pgp-sftp-proxy/temp.json \
 && /usr/bin/pgp-sftp-proxy -c /usr/local/pgp-sftp-proxy/temp.json
