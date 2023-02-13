# Roadmap

-   Implement everything in Go - in progress
    -   Setting up bots via SSH and BLE should have a Go implementation
    -   Initial setup should also be done in Go. Certificates, ports, and model downloads configurable via web interface

-   Tokens which are unique, do expire, and do refresh - DONE - 12/22/22
    -   Retain backwards compatibility
        -   Global GUID code still in there
    -   Jdocs properly keep DocVersion and FmtVersion
    -   Unique GUID per bot
        -   Currently, there is one global GUID used for all bots
        -   When a GUID is generated, add it to the sdk_config.ini file
            -   If it isn't there, generate the file
    -   Reason for difficulty: the official servers have access to every bot's factory certificate, and are able to get the serial number of the requester bot by checking the CN of the certificate. We have to work around that because we don't have access to those certs
    -   Other things changed due to this: removed INI-to-JSON which took the sdk_config.ini file and put the GUID and stuff from that into wire-pod's botinfo JSON. It wasn't that useful because we don't have access to the token hashes

-   Implement face recognition settings in sdkApp - DONE- 12/24/22
    -   Change name of someone
    -   Delete a face
    -   Maybe initiate a recognition

-   Implement photo export in sdkApp

-   Refactor - always in progress
    -   Goal: Make it simple to implement your own STT engine, clean things up
    -   Separate more things into their own components
        -   speechrequest, wirepod_coqui(etc...), ttr(text-to-response)

-   Speech-to-intent support - IN PROGRESS
    -   Status: Implemented, but still experimental and must be set up manually (not in setup.sh)
    -	I would like to implement Picovoice Rhino because it works on Vector's CPU
    -	The current code is designed to accomodate just speech-to-text engines

-   Streaming to houndify - DONE - 1/8/22
    -   Status: Implemented. Audio is streamed to Houndify and VAD is used to figure out when to stop the stream. Response comes in surprisingly fast
    -   New findings after implementing: After experimenting with Houndify's dashboard, it seems to be what DDL is piping every single cloud voice request into. It has "custom commands" which they likely programmed every intent into, and it falls back to a normal knowledge graph response if it doesn't match, which is how the "intent graph" feature works. Also, their direct speech-to-text is super fast (after my VAD modification) and I may implement it as an option (NOTE: 1/8/22 - it is now implemented as an experimental option). Though it is a little difficult to describe how to set it up in the dashboard
    -   Why (pre-impl)
        -	Currently, VAD is used to detect the end of speech and a big chunk of data is sent all at once to the Houndify servers
        -	Streaming is supported by their servers and I would like to implement this
	    -   This would make requests a lot faster

-   Dynamic config - COMPLETE
    -   Rather than using environment variables for API configs, put the variables into a global, updatable struct. This can be exported to JSON
    -   Will allow the changing of KG or weather provider without restarting wirepod, or even the voice processor itself

-   Implement wire-pod status page
    -   Status: logs work
    -   Should include:
        -   Bot serial numbers
        -   Saved bot targets (IP addresses)
        -   How long it has been since last connCheck per bot
    -   May end up being a wire-prod-pod thing only
