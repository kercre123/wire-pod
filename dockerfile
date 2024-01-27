FROM ubuntu

WORKDIR /app
COPY . /app

RUN chmod +x /app/setup.sh && apt-get update && apt-get install -y dos2unix && dos2unix /app/setup.sh

RUN ["/bin/sh", "-c", "STT=vosk ./setup.sh"]

RUN chmod +x /app/chipper/start.sh && dos2unix /app/chipper/start.sh

CMD ["/bin/sh", "-c", "./chipper/start.sh"]