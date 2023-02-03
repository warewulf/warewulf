# Warewulf API Daemon

wwapid is a grpc service serving the Warewulf API. For v1, the intent is to serve the same interface as wwctl serves. For this preview PR, we are serving much of ```wwctl node``` and ```wwctl container```.

Initial security is by mTLS. For development we are generating our own keys. A good tutorial for this is here: https://www.handracs.info/blog/grpcmtlsgo/

The configuration file is wwapid.conf.