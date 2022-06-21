# wire-pod

This repo contains a custom Vector escape pod made from [chipper](https://github.com/digital-dream-labs/chipper) and [vector-cloud](https://github.com/digital-dream-labs/vector-cloud).

## Program Descriptions

`chipper` - Chipper is a program used on Digital Dream Lab's servers which takes in a Vector's voice stream, puts it into a speech-to-text processor, and spits out an intent. This is also likely used on the official escape pod. This repo contains an older tree of chipper which does not have the "intent graph" feature (it caused an error upon every new stream), and it now has a working voice processor.

`vector-cloud` - Vector-cloud is the program which runs on Vector himself which uploads the mic stream to a chipper instance. This repo has an older tree of vector-cloud which also does not have the "intent graph" feature and has been modified to allow for a custom CA cert.

## System Requirements

A CPU with AVX support is required. Check [https://en.wikipedia.org/wiki/Advanced_Vector_Extensions#CPUs_with_AVX](https://en.wikipedia.org/wiki/Advanced_Vector_Extensions#CPUs_with_AVX)

## Configuring, Building, Installing

NOTE: This only works with OSKR-unlocked, Dev-unlocked, or Whiskey robots.

### Linux

(This currently only works on Arch or Debian-based Linux)

```
git clone https://github.com/kercre123/wire-pod.git
cd wire-pod
sudo ./setup.sh

# You should be able to just press enter for all of the settings
```

Now install the files created by the script onto the bot:

`sudo ./setup.sh scp <vectorip> <path/to/key>`

Example:

`sudo ./setup.sh scp 192.168.1.150 /home/wire/id_rsa_Vector-R2D2`

If you are on my custom software (WireOS), you do not have to provide an SSH key,

Example:

`sudo ./setup.sh scp 192.168.1.150`

The bot should now be configured to communicate with your server.

To start chipper, run:

```
cd chipper
sudo ./start.sh
```

### Windows

1. Install WSL (Windows Subsystem for Linux)
	- Open Powershell
	- Run `wsl --install`
	- Reboot the system
	- Run `wsl --install -d Ubuntu-20.04`
	- Open up Ubuntu 20.04 in start menu and configure it like it says.
2. Find IP address
	- Open Powershell
	- Run `ipconfig`
	- Find your computer's IPv4 address and note it somewhere. It usually starts with 10.0. or 192.168.
3. Install wire-pod
	- Follow the Linux instructions from above
	- Enter the IP you got from `ipconfig` earlier instead of the one provided by setup.sh
	- Use the default port and do not enter a different one
4. Setup firewall rules
	- Open Powershell
	- Run `Set-ExecutionPolicy`
	- When it asks, enter `Bypass`
	- Download [this file](https://wire.my.to/wsl-firewall.ps1)
	- Go to your Downloads folder in File Explorer and Right Click -> Run as administrator


After all of that, try a voice command.

## Status

OS Support:

- Arch
- Debian/Ubuntu/other APT distros

Architecture support:

- amd64/x86_64
- arm64/aarch64
- arm32/armv7l

Things wire-pod has worked on:

- Raspberry Pi 4B+ 4GB RAM with Raspberry Pi OS
	- Doesn't matter if it is 32-bit or 64-bit
- Raspberry Pi 4B+ 4GB RAM with Manjaro 22.04
- Nintendo Switch with L4T Ubuntu
- Desktop with Ryzen 5 3600, 16 GB RAM with Ubuntu 22.04
- Laptop with mobile i7
- Android Devices
	- Pixel 4, Note 4, Razer Phone, Oculus Quest 2, OnePlus 7 Pro, Moto G6, Pixel 2
	- [Termux](https://github.com/termux/termux-app) proot-distro: Use Ubuntu, make sure to use a port above 1024 and not the default 443.
	- Linux Deploy: Works stock, just make sure to choose the arch that matches your device in settings. Also use a bigger image size, at least 3 GB.

General notes:

- If the architecture is AMD64, the text is processed 4 times so longer phrases get processed fully. Text is only processed once on arm32/arm64 for speed and reliability.
	- If you are running ARM and you feel like your system is fast enough for regular STT processing, 'sudo rm -f ./chipper/slowsys'
	- If you are running AMD64 and you feel like your system is too slow for regular STT processing, `touch ./chipper/slowsys'
- If you get this error when running chipper, you are using a port that is being taken up by a program already: `panic: runtime error: invalid memory address or nil pointer dereference`
	- Run `./setup.sh` with the 5th and 6th option to change the port, you will need to push files to the bot again.
- If you want to disable logging from the voice processor, recompile chipper with `debugLogging` in ./chipper/pkg/voice_processors/noop/intent.go set to `false`.
- Some weather conditions don't match the bot's internal response map, so sometimes Vector won't be able to give you the weather until I make my own response map.
- You have to speak a little slower than normal for Coqui STT to understand you.

Known issues:

- ARM processing is slow until I find a good way to deal with the end of speech on (comparatively) slow hardware.

Current implemented actions:

- Good robot
- Bad robot
- Change your eye color
- Change your eye color to <color>
	- blue, purple, teal, green, yellow
- How old are you
- Start exploring ("deploring" works better)
- Go home (or "go to your charger")
- Go to sleep
- Good morning
- Good night
- What time is it
- Goodbye
- Happy new year
- Happy holidays
- Hello
- Sign in alexa
- Sign out alexa
- I love you
- Move forward
- Turn left
- Turn right
- Roll your cube
- Pop a wheelie
- Fistbump
- Blackjack (say yes/no instead of hit/stand)
- Yes (affirmative)
- No (negative)
- What's my name
- Take a photo
- Take a photo of me
- What's the weather
	- Requires API setup
- What's the weather in <location>
	- Requires API setup
- Im sorry
- Back up
- Come here
- Volume down
- Be quiet
- Volume up
- Look at me
- Set the volume to <volume>
	- High, medium high, medium, medium low, low
- Shut up
- My name is <name>
- I have a question
	- Placeholder text
- Set a timer for <time> seconds
- Set a timer for <time> minutes
- Check the timer
- Stop the timer
- Dance
- Pick up the cube
- Fetch the cube
- Find the cube
- Do a trick
- Record a message for <name>
	- Enable `Messaging` feature in webViz Features tab
- Play a message for <name>
	- Enable `Messaging` feature in webViz Features tab

## Credits

- [Digital Dream Labs](https://github.com/digital-dream-labs) for saving Vector and for open sourcing chipper which made this possible
- [dietb](https://github.com/dietb) for rewriting chipper and giving tips
- [GitHub Copilot](https://copilot.github.com/) for being awesome
