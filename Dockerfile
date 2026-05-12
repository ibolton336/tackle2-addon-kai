FROM registry.access.redhat.com/ubi9/go-toolset:latest as addon
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
RUN make cmd

FROM golang:1.22 AS pallet-builder
RUN git clone https://github.com/djzager/pallet.git /src/pallet && \
    cd /src/pallet && \
    CGO_ENABLED=0 go build -o /usr/local/bin/pallet .

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
RUN echo -e "[centos9]" \
 "\nname = centos9" \
 "\nbaseurl = http://mirror.stream.centos.org/9-stream/AppStream/\$basearch/os/" \
 "\nenabled = 1" \
 "\ngpgcheck = 0" > /etc/yum.repos.d/centos.repo
RUN microdnf -y install \
 glibc-langpack-en \
 openssh-clients \
 subversion \
 git \
 tar \
 && curl -fsSL https://github.com/block/goose/releases/latest/download/goose-linux-x86_64 -o /usr/bin/goose \
 && chmod +x /usr/bin/goose
RUN sed -i 's/^LANG=.*/LANG="en_US.utf8"/' /etc/locale.conf
ENV LANG=en_US.utf8
RUN echo "addon:x:1001:1001:addon user:/addon:/sbin/nologin" >> /etc/passwd
RUN echo -e "StrictHostKeyChecking no" \
 "\nUserKnownHostsFile /dev/null" > /etc/ssh/ssh_config.d/99-konveyor.conf
ENV HOME=/addon ADDON=/addon
WORKDIR /addon
ARG GOPATH=/opt/app-root
COPY --from=addon $GOPATH/src/bin/addon /usr/bin
COPY --from=pallet-builder /usr/local/bin/pallet /usr/bin/pallet
COPY skills/ /addon/skills/
ENTRYPOINT ["/usr/bin/addon"]
