We need an API for Warewulf in order to automate around it. The initial version of the API will look a lot like wwctl. wwctl will call the API by direct function call so that we do not need to maintain separate code paths, nor is the API required.

wwctl calls the API via direct function call, but we can change that to a gprc or REST call to a Warewulf server in the future.

The API currently contains:
wwapid WareWulf API Daemon. A grpc server with mTLS auth.
wwapic WareWulf API Client. A grpc client with mTLS auth. It's just a sample that reads the version from the server.
wwapird WareWulf API Rest Daemon. A http REST reverse proxy to wwapid.
There are also sample insecure and secure curls to test with.

Implemented Functionality:
wwctl node add
wwctl node delete
wwctl node list
wwctl node set
wwctl node status
wwctl container build
wwctl container delete
wwctl container import
wwctl container list
wwctl container show
wwctl container copy

Some notes on the files:

Logic that was in wwctl has moved to warewulf/internal/pkg/api. wwctl just calls the API. wwctl functionality is unchanged.
Everything under the /google directory is copied google proto code to stand up wwapird. It's a bit strange to copy in the code, but that is currently how it's done.
Everything in warewulf/internal/pkg/api/routes/wwapiv1 is generated code.

There are some loose ends here, such as no service installers. I just ran off the command line for development.
The Makefile could use improvement. I'm building by make clean setup proto all build wwapid wwapic wwapird ; echo $?
I copied the configs from warewulf/etc to /usr/local/etc/warewulf manually.
