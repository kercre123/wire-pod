# wire-pod with Wyoming speech-to-text

This branch adds a `wyoming` STT provider: instead of the built-in Vosk,
wire-pod sends the robot's audio to an external
[Wyoming](https://github.com/OHF-Voice/wyoming) speech-to-text server, the same
kind Home Assistant uses (wyoming-faster-whisper, wyoming-vosk, GigaAM via
onnx-asr...). One model can then serve both the robot and Home Assistant, and
wire-pod itself needs only a few dozen MB of RAM instead of gigabytes for a
large Vosk model.

What the branch changes:

- `chipper/pkg/wirepod/stt/wyoming` and `chipper/cmd/wyoming`: the provider.
  It collects audio until the end of speech, sends `transcribe`,
  `audio-start`/`audio-chunk`/`audio-stop` and waits for `transcript`.
  wire-pod still starts if the server is down.
- Web interface, Server Settings -> STT Language: server address, model,
  language sent to the server, timeout and a **Test connection** button that
  lists the server's models.
- The chosen language is kept for providers that support it; before, any
  provider other than vosk/whisper was forced to `en-US` intents.
- Intent matching:
  - "ё" is treated as "е" when comparing (the recognized text itself is
    passed on unchanged);
  - short built-in keyphrases (3 letters or less, e.g. "да", "нет") match
    only as whole words, so "да" no longer matches "давай";
  - an exact built-in phrase is matched before custom intents and plugins,
    whose utterances are matched as substrings.

## Build

The packaged build uses `debian/pod` from
[kercre123/WirePod](https://github.com/kercre123/WirePod); `WirePod.patch`
makes it start the Wyoming provider instead of Vosk. As root on a Linux host:

```sh
wyoming-build/build.sh /opt/build
```

Go is downloaded into the work directory; only `pkg-config` and `libopus-dev`
are installed system-wide. The script prints the glibc version the binary
needs: it must not be newer than on the host that runs wire-pod (a build on
Ubuntu 24.04 needs 2.34). libvosk is not linked.

## Install over a packaged (.deb) wire-pod

```sh
# backup for rollback
tar czf /root/wire-pod-backup.tgz --exclude=etc/wire-pod/wire-pod/vosk -C / \
  etc/wire-pod/webroot etc/wire-pod/intent-data etc/wire-pod/weather-map.json etc/wire-pod/wire-pod
install -m 0755 out/wire-pod-wyoming /usr/bin/wire-pod-wyoming
# the old webroot does not work with the new API guard, replace it together
cd /etc/wire-pod && for d in webroot intent-data weather-map.json; do mv $d $d.vosk; done
cp -r /opt/build/out/{webroot,intent-data,weather-map.json} /etc/wire-pod/
```

`/etc/systemd/system/wire-pod.service.d/wyoming.conf`:

```ini
[Service]
ExecStart=
ExecStart=/usr/bin/wire-pod-wyoming
# used until the address is saved in the web interface
Environment=WYOMING_STT_ADDRESS=192.168.1.10:10300
# public domains are rejected by the API guard unless listed
#Environment=WIREPOD_ALLOWED_HOSTS=wirepod.example.com
#Environment=WIREPOD_ALLOWED_ORIGINS=https://wirepod.example.com
```

```sh
systemctl daemon-reload && systemctl restart wire-pod
journalctl -u wire-pod -f
# expect "Initiating wyoming voice processor with language ..."
# and "Wyoming STT: <address> (<program>: <models>)"
```

Other environment variables: `WYOMING_STT_MODEL`, `WYOMING_STT_LANGUAGE`,
`WYOMING_STT_TIMEOUT` (seconds); values saved in the web interface win.

Rollback: delete `wyoming.conf`, move the `*.vosk` copies back, then
`systemctl daemon-reload && systemctl restart wire-pod`. The original
`/usr/bin/wire-pod` starts with Vosk again, as long as its model is still in
`/etc/wire-pod/wire-pod/vosk`.

If the Wyoming server goes down, the robot cannot recognize speech at all.
