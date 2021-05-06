# DFS IDEAS

## underlying disc ops /fs apis
* we're probably gonna use [FUSE](https://github.com/libfuse/libfuse) and its implementation in GO or C
* Filesystem in Userspace , just use their API's for hardcore R/W ; all data stored as blob on FS .

## BLOB storage mechanism
* we are going to implement a FS , with Ideas taken from Facebook's [Finding  a needle in Haystack](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf) architecture and [f4 Facebook's Warm BLOB Storage System](https://www.usenix.org/system/files/conference/osdi14/osdi14-paper-muralidhar.pdf) as they seem to be state of the art from initial findings on our part.

## RPC
* could use HTTP REST , but thinking of going [gRPC](https://grpc.io/)  , and thus have to use [Protocol Buffers](https://github.com/protocolbuffers/protobuf)
* ```protoc``` compiles protobuf to any language , thus RPC is language agnostic.

## Fault Tolernace / Consensus 
* Mostly gonna use [Raft](https://web.stanford.edu/~ouster/cgi-bin/papers/raft-atc14) , depending on guide's advice we'll see if we're gonna implement it from ground up or modify an existing implementation for our use (LogEntry RPC needs to work with Haystack)
