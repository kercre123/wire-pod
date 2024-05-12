FROM ubuntu


# *** PACKAGE INSTALLS ***
# install packages before copying in source so we don't do this on every source change
RUN apt-get update && \
  apt-get install -y dos2unix avahi-daemon avahi-autoipd
#
# setup.sh is standalone (IIUC upon cursory inspection) at least for debian/aarch64 + vosk purposes
# setup.sh is more or less further `apt install...` + golang install + vosk install + gen chipper/source.sh... thus run this as early as possible as it is not likely to change (as much) as the source code
COPY setup.sh /setup.sh
# PRN part of what setup.sh does is to install golang and other deps... why not extract those deps here for the Dockerfile and not use setup.sh inside the container build? AND why not find a golang base image to build on top of instead of installing it here?
RUN dos2unix /setup.sh && mkdir /chipper && mkdir /vector-cloud
RUN ["/bin/sh", "-c", "STT=vosk IMAGE_BUILD=true SETUP_STAGE=getPackages ./setup.sh"]

# so we can install go deps (IIUC for vosk install in setup.sh/getSTT)
COPY ./chipper/go.sum ./chipper/go.mod /chipper/
# FYI setup.sh generates /chipper/source.sh which sets env vars only
RUN ["/bin/sh", "-c", "STT=vosk IMAGE_BUILD=true SETUP_STAGE=getSTT ./setup.sh"]
# *** END PACKAGE INSTALLS ***

# *** go download deps ***
#
# PRN build + cache deps:  # RUN --mount=type=cache,target=/root/go/pkg/mod cd /chipper && go mod download
#
COPY ./vector-cloud/go.sum ./vector-cloud/go.mod /vector-cloud/
RUN cd /chipper && go mod download
# vector-cloud deps are normally downloaded after initial web server Submit Settings page:
RUN cd /vector-cloud && go mod download
# PRN IIUC vector-cloud has some vosk model download/setup, can I add it to this image build too?
# *** END go download deps ***

# PRN copy specific files only: # COPY chipper images vector-cloud /
COPY . .
RUN dos2unix /chipper/start.sh
# TODO do we really need dos2unix? can't we use editorconfig or something else to enforce line endings? and/or force git checkout to have LF endings always? SAME with setup.sh above too
RUN ["/bin/sh", "-c", "DO_GO_BUILD=true ./chipper/start.sh"]
# TODO go build chipper&vector-cloud - bake into this image (perhaps modify start.sh to do this (sep build param) or just inline go build here)
CMD ["/bin/sh", "-c", "./chipper/start.sh"]