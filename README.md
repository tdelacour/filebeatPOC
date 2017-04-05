# Proof of concept for using filebeat as a library

Simple POC to demonstrate using filebeat's `crawler` and `registrar` as
external packages. These run alongside a custom `spooler`/`publisher` that 
batches log file events together and sends them to the http server that is also
in this repository.

## Setup
- Place this repository under $GOPATH/src/github.com/tdelacour/
- Clone my fork of beats (github.com/tdelacour/beats) and place under
$GOPATH/src/github.com/elastic/

## To run
```
make
make runServer
```
in another tab
```
make runFilebeat
```
