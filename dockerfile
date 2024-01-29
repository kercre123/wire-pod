FROM ubuntu


COPY . .

RUN chmod +x /setup.sh && apt-get update && apt-get install -y dos2unix && dos2unix /setup.sh && apt-get install -y avahi-daemon avahi-autoipd

RUN ["/bin/sh", "-c", "STT=vosk ./setup.sh"]

RUN chmod +x /chipper/start.sh && dos2unix /chipper/start.sh

CMD ["/bin/sh", "-c", "./chipper/start.sh"]