# Proof of concept for using filebeat as a library

Simple POC to demonstrate using filebeat's `crawler` and `registrar` as
external packages. These run alongside a custom `spooler`/`publisher` that 
batches log file events together and sends them to the http server that is also
in this repository.

## Setup
- Place this repository under $GOPATH/src/github.com/tdelacour/
- Clone my fork of beats (https://github.com/tdelacour/beats.git) and place under
$GOPATH/src/github.com/elastic/

## To run
```
make
make runServer
```
in another tab
```
make glob='/path/to/logfiles' runFilebeat
```
For example, if a couple of processes are logging to /var/log/log1 and /var/log/log2
you would run
```
make glob='/var/log/log*' runFilebeat
```
