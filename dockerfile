FROM ubuntu


# install packages before copying in source so we don't do this on every source change
RUN apt-get update && \
  apt-get install -y dos2unix avahi-daemon avahi-autoipd

COPY . .

RUN apt-get update && dos2unix /setup.sh

RUN ["/bin/sh", "-c", "STT=vosk ./setup.sh"]

RUN dos2unix /chipper/start.sh

CMD ["/bin/sh", "-c", "./chipper/start.sh"]