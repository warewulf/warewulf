# Warewulf API Rest Daemon

wwapird is a client of wwapid serving a REST interface for the Warewulf API. For v1, the intent is to serve the same interface as wwctl serves. For this preview PR, we are serving much of ```wwctl node``` and ```wwctl container```.

Initial security is by mTLS. For development we are generating our own keys. A good tutorial for this is here: https://www.handracs.info/blog/grpcmtlsgo/

The configuration file is wwapird.conf.

Sample cURLs with and without mTLS are in curl.sh. This file is just a scratchpad for examples.