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
# FYI setup.sh generates /chipper/source.sh: contains env vars for STT_SERVICE & DEBUG_LOGGING for vosk setup.sh:
RUN ["/bin/sh", "-c", "STT=vosk IMAGE_BUILD=true SETUP_STAGE=getSTT ./setup.sh"]
# *** END PACKAGE INSTALLS ***

# TODO figure out if anything gets clobbered that was created by setup.sh (i.e. ./chipper/source.sh which is created by setup.sh)
# TODO can we only copy:
#  COPY chipper images /
#    IIAC docker doesn't need update.sh (b/c instead just rebuild image, IIAC)
#    IIUC vector-cloud isn't part of the server (gh issue mentions possibly removing it - is it part of custom ep firmeware that runs on vector?)
COPY . .
RUN dos2unix /chipper/start.sh
# TODO do we really need dos2unix? can't we use editorconfig or something else to enforce line endings? and/or force git checkout to have LF endings always? SAME with setup.sh above too

# TODO step 1 - add go get/install for runtime modules into a last layer
#   setup 2 - ascertain if I can move go get/install earlier in the Dockerfile (if that has any benefit)
CMD ["/bin/sh", "-c", "./chipper/start.sh"]