FROM ubuntu


# install packages before copying in source so we don't do this on every source change
RUN apt-get update && \
  apt-get install -y dos2unix avahi-daemon avahi-autoipd

# setup.sh is standalone (IIUC upon cursory inspection) at least for debian/aarch64 + vosk purposes
COPY setup.sh /setup.sh
RUN dos2unix /setup.sh && mkdir /chipper
RUN ["/bin/sh", "-c", "STT=vosk ./setup.sh"]

# TODO figure out if anything gets clobbered that was created by setup.sh (i.e. ./chipper/source.sh which is created by setup.sh)
COPY . .

RUN dos2unix /chipper/start.sh

CMD ["/bin/sh", "-c", "./chipper/start.sh"]