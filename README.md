# sfu-load-test : ion-sfu load testing tool

## Overview [![GoDoc](https://godoc.org/github.com/smf8/sfu-load-test?status.svg)](https://godoc.org/github.com/smf8/sfu-load-test)

In my path of working around with [pion/webrtc](https://github.com/pion/webrtc/) library, I came across a fantastic repo called [ion](https://github.com/pion/ion) which concentrated on a scalable video conference system. A major part of ion is it's [**S**elective **F**orward **U**nit](https://webrtcglossary.com/sfu/) called [`ion-sfu`](https://github.com/pion/ion-sfu). As for the time of creating this repo, there wasn't any load testing tool for `ion-sfu` itself (there is a repo for ion's stress testing called [`ion-load-tool`](https://github.com/pion/ion-load-tool)). I decided to create this repo for it's gRPC client.



**Note**: ion-sfu is under haevy code base changes and this repo might not work in the future. 

## Install

```go
go get github.com/smf8/sfu-load-test
```

Same as `ion-load-tool` `.ivf` and `.webm` video formats are supported with `VP8` or `VP9` codecs.

You can start `ion-sfu` from it's docker registry and monitor the test with `docker stats`.

You can also edit what subscribers can do with received packets in `subscriber/subscriber.go` and `handleTrack(tr *webrtc.Track) ` function.

## Example

possible flags:

- `sub`: determine number of subscriber clients
- `pub`: determine number of publisher clients
- `server`: gRPC server and port in format `address:port`
- `sid`: which session in SFU to use
- `file`: video file path.

```shell
make all

#For a single publisher
bin/sfu-load-test -pub 1 -file ./video-file.webm -sid 10

#For a single subscriber
bin/sfu-load-test -sub 1 -sid 10
```
