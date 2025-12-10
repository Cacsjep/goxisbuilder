ARG ARCH=aarch64
ARG VERSION=12.2.0
ARG UBUNTU_VERSION=24.04
ARG REPO=axisecp
ARG SDK=acap-native-sdk 
FROM ${REPO}/${SDK}:${VERSION}-${ARCH}-ubuntu${UBUNTU_VERSION}

ARG ARCH
ARG VERSION
ARG SDK_LIB_PATH_BASE=/opt/axis/acapsdk/sysroots/${ARCH}/usr
ARG APP_DIR=/opt/goaxis/
ARG GOLANG_VERSION=1.25.3
ARG APP_NAME=app
ARG APP_MANIFEST=
ARG GO_ARCH=arm64
ARG GO_ARM=
ARG IP_ADDR= 
ARG PASSWORD= 
ARG START=
ARG DONT_COPY=
ARG INSTALL=
ARG FILES_TO_ADD_TO_ACAP=
ARG GO_APP=test
ARG GO_BUILD_TAGS=
ARG ENABLE_UPX=YES

ENV GOPATH="/go" \
    PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}" \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=${GO_ARCH} \
    GOARM=${GO_ARM} \
    APP_NAME=${APP_NAME} \
    ACAP_FILES=${FILES_TO_ADD_TO_ACAP} \
    MANIFEST=${APP_MANIFEST} \
    GO_APP=${GO_APP} \
    GO_BUILD_TAGS=${GO_BUILD_TAGS} \
    ENABLE_UPX=${ENABLE_UPX}


RUN apt-get update && apt-get install -y upx-ucl




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

# Copy go.mod and go.sum files to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . ${APP_DIR}
WORKDIR ${APP_DIR}
RUN python generate_makefile.py ${APP_NAME} ${GO_APP} ${APP_MANIFEST}
WORKDIR ${APP_DIR}/${GO_APP}
RUN . /opt/axis/acapsdk/environment-setup* && \
    make build && \
    if [ "$ENABLE_UPX" = "YES" ]; then \
        echo "Compressing binary with UPX..."; \
        upx --best --lzma ${APP_NAME} || echo "UPX failed, continuing with uncompressed binary"; \
    fi && \
    acap-build . ${ACAP_FILES} || (echo "acap-build error" && exit 1)

# Install ACAP only if INSTALL=YES
RUN . /opt/axis/acapsdk/environment-setup* && if [ "$INSTALL" = "YES" ]; then eap-install.sh ${IP_ADDR} ${PASSWORD} install || (echo "acap-build error install" && exit 1); fi

# Start ACAP only if START=YES
RUN . /opt/axis/acapsdk/environment-setup* && if [ "$START" = "YES" ]; then eap-install.sh start || (echo "acap-build error start" && exit 1); fi

#----------------------------------------------------------------------------
# Conditional Copy out the eap file
#----------------------------------------------------------------------------
RUN if [ "$DONT_COPY" = "NO" ]; then \
  mkdir /opt/build && \
  mv *.eap /opt/build && \
  cd /opt/build && \
  for file in *.eap; do \
        mv "$file" "${file%.eap}_sdk_${VERSION}.eap"; \
  done; \
fi
