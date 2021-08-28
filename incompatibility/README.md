# Building and Running Incompatibility scanning

In order to build the gnostic-grpc plugin make sure you have the pre-requisities as defined in the general [README](https://github.com/google/gnostic-grpc/blob/master/README.md) and then run the shell script

    ./plugin-creation.sh

An IncompatibilityReport which is defined [here](https://github.com/google/gnostic-grpc/blob/master/incompatibility/incompatibility-report.proto) is not very informative but provides a classification for the incompatibilty as well as a severity level. Run the command with an appropriate file path to an OpenAPI document and an output location

    gnostic --grpc-out=report=1:<output> <document>
An FileDescriptiveReport  which is defined [here](https://github.com/google/gnostic-grpc/blob/master/incompatibility/incompatibility-report.proto), can be useful when singular file focused incompatibility information is needed. This report pairs incompatibilities with file position and a string detailing the incompatibility. To get such a report run the command with an appropriate file path to an OpenAPI document and an output location

    gnostic --grpc-out=report=2:<output> <document>