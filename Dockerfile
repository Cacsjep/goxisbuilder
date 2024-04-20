ARG ARCH=aarch64
ARG VERSION=1.13
ARG UBUNTU_VERSION=22.04
ARG REPO=axisecp
ARG SDK=acap-native-sdk 
FROM ${REPO}/${SDK}:${VERSION}-${ARCH}-ubuntu${UBUNTU_VERSION}

ARG ARCH
ARG VERSION
ARG SDK_LIB_PATH_BASE=/opt/axis/acapsdk/sysroots/${ARCH}/usr
ARG APP_DIR=/opt/goaxis/
ARG GOLANG_VERSION=1.22.1
ARG APP_NAME=app
ARG APP_MANIFEST=
ARG GO_ARCH=arm64
ARG GO_ARM=
ARG IP_ADDR= 
ARG PASSWORD= 
ARG START=
ARG INSTALL=
ARG FILES_TO_ADD_TO_ACAP=
ARG GO_APP=test

ENV GOPATH="/go" \
    PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}" \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=${GO_ARCH} \
    GOARM=${GO_ARM} \
    APP_NAME=${APP_NAME} \
    ACAP_FILES=${FILES_TO_ADD_TO_ACAP} \
    MANIFEST=${APP_MANIFEST} \
    GO_APP=${GO_APP}

RUN mkdir ${APP_DIR}

#-------------------------------------------------------------------------------
# Golang build
#-------------------------------------------------------------------------------
RUN curl -fsSL "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" -o golang.tar.gz \
    && tar -C /usr/local -xzf golang.tar.gz \
    && rm golang.tar.gz
RUN mkdir -p "${GOPATH}/src" "${GOPATH}/bin" "${GOPATH}/pkg" \
    && chmod -R 777 "${GOPATH}"

#-------------------------------------------------------------------------------
# ACAP Build
#-------------------------------------------------------------------------------
COPY . ${APP_DIR}
WORKDIR ${APP_DIR}
RUN python generate_makefile.py ${APP_NAME} ${GO_APP} ${APP_MANIFEST}
WORKDIR ${APP_DIR}/${GO_APP}
RUN . /opt/axis/acapsdk/environment-setup* && \
    acap-build . ${ACAP_FILES} && \
    if [ "${INSTALL}" = "YES" ]; then eap-install.sh ${IP_ADDR} ${PASSWORD} install; fi && \
    if [ "${START}" = "YES" ]; then eap-install.sh start; fi

#-------------------------------------------------------------------------------
# Create output directory, we copy files from eap to host
#-------------------------------------------------------------------------------
RUN mkdir /opt/build
RUN mv *.eap /opt/build
RUN cd /opt/build && \
    for file in *.eap; do \
        mv "$file" "${file%.eap}_sdk_${VERSION}.eap"; \
    done

