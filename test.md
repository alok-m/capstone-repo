# File System CLI Prototype

## Build instructions

run 
``` 
sh build.sh
```

## commands 

* Volume create
```
./fs_test volume create <volume directory>
```
* Volume delete
```
./fs_test volume delete <volume directory>
```
* File write 
```
./fs_test file write <fid> <volume directory> <input file>
```
* File read
```
./fs_test file read <fid> <volume directory> <output File>
```
* File delete
```
./fs_test file delete <fid> <volume directory>
```