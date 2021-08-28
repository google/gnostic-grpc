# Running Analysis

This analysis tool internally invokes gnostic so please follow the directions outline in the general [README](https://github.com/google/gnostic-grpc/blob/master/README.md).

An ApiSetIncompatibility is an analysis object which contains incompatibility information over a set of OpenAPI documents defined [here](https://github.com/google/gnostic-grpc/blob/master/incompatibility/incompatibility-report.proto). In order to run this tool against a directory run the command with the specified directory


    go run analysis/analysis.go <directory>