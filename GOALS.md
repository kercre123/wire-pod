# Roadmap

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

-   Refactor - IN PROGRESS - 12/27/22
    -   Goal: Make it simple to implement your own STT engine, clean things up
    -   Separate more things into their own components
        -   speechrequest, wirepod_coqui(etc...), ttr(text-to-response)

-   Implement wire-pod status page
    -   Should include:
        -   Bot serial numbers
        -   Saved bot targets (IP addresses)
        -   How long it has been since last connCheck per bot
    -   May end up being a wire-prod-pod thing only
