#!/bin/bash

if [[ -f armarch ]]; then
  touch slowsys
fi

botNum=$1

if [[ $2 == "kg" ]]; then
  fileSuffix="kg"
else
  fileSuffix=""
fi

function sstCmd() {
STT_TFLITE_DELEGATE=gpu ./stt/stt --model ./stt/model.tflite --scorer ./stt/large_vocabulary.scorer --audio /tmp/${botNum}voice${fileSuffix}.wav
}

function ffmpegCmd() {
ffmpeg -y -i /tmp/${botNum}voice${fileSuffix}.ogg /tmp/${botNum}voice${fileSuffix}.wav
}

function doSttSlow() {
sleep 1.5
cd ../
rm -rf /tmp/${botNum}voice${fileSuffix}.wav
ffmpegCmd
sstCmd > /tmp/${botNum}utterance1${fileSuffix}
touch /tmp/${botNum}sttDone${fileSuffix}
sleep 0.5
rm -f /tmp/${botNum}utterance*${fileSuffix}
rm -f /tmp/${botNum}voice${fileSuffix}.wav
rm -rf /tmp/${botNum}voice${fileSuffix}.ogg
rm -f /tmp/${botNum}sttDone${fileSuffix}
}

function doSttFast() {
sleep 0.8
cd ../
rm -rf /tmp/${botNum}voice.wav${fileSuffix}
ffmpegCmd
sstCmd > /tmp/${botNum}utterance1${fileSuffix}
sleep 0.5
ffmpegCmd
sstCmd > /tmp/${botNum}utterance2${fileSuffix}
ffmpegCmd
sstCmd > /tmp/${botNum}utterance3${fileSuffix}
ffmpegCmd
sstCmd > /tmp/${botNum}utterance4${fileSuffix}
touch /tmp/${botNum}sttDone${fileSuffix}
sleep 0.5
rm -rf /tmp/${botNum}utterance*${fileSuffix}
rm -rf /tmp/${botNum}voice${fileSuffix}.wav
rm -rf /tmp/${botNum}voice${fileSuffix}.ogg
rm -rf /tmp/${botNum}sttDone${fileSuffix}
}

if [[ -f slowsys ]]; then
  doSttSlow &
else
  doSttFast &
fi
