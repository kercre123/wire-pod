# chipper

This repo is a clone of [chipper](https://github.com/digital-dream-labs/chipper)

It also contains a clone of [vector-cloud](https://github.com/digital-dream-labs/vector-cloud), but with some modifications so a custom cert can be used easily.

## Changes

This is from an older tree of both chipper and vector-cloud. It appears chipper became non-functional after intent-graph was added (errored out upon NewStream). 

The original vector-cloud seemed to have the escape pod cert appended to the pool in the jdocs component, rather than the voice component.

The cert is now located in vector-cloud/cloud/main.go. I have included some example certs which point to ankitestankites.mooo.com

There is no speech-to-text implementation right now. It just saves the audio to /tmp/test.ogg and gives the robot the praise intent

## Build

`make build`
