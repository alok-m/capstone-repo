package storage

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	TruncateSize  uint64 = 1 << 30            //1GB
	MaxVolumeSize uint64 = 512 * TruncateSize //512GB
)

// Volume definition
type Volume struct {
	vid       uint64
	WriteAble bool

	index    Index   // CRUD abstraction of the volume
	status   *Status // locations of free blocks stored
	dataFile *os.File
	mutex    sync.Mutex
}

func NewVolume(dir string, vid uint64) (v *Volume, err error) {
	path := filepath.Join(dir, strconv.FormatUint(vid, 10)+".data")
	v = new(Volume)
	v.vid = vid
	v.dataFile, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		if os.IsPermission(err) {
			v.dataFile, err = os.OpenFile(path, os.O_RDONLY, 0)
			if err != nil {
				return nil, err
			}
			v.WriteAble = false
		} else {
			return nil, err
		}
	}
	v.WriteAble = true

	v.index, err = NewLevelDBIndex(dir, vid)
	if err != nil {
		return nil, err
	}

	v.status, err = NewStatus(dir, vid)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Volume) Get(fid uint64) (*File, error) {
	//v.rwMutex.RLock()
	//defer v.rwMutex.RUnlock()

	fi, err := v.index.Get(fid)
	if err == nil {
		return &File{DataFile: v.dataFile, Info: fi}, nil
	}
	return nil, err

}

func (v *Volume) Delete(fid uint64, fileName string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	file, err := v.Get(fid)
	if err != nil {
		return err
	} else if file.Info.FileName != fileName {
		return errors.New("filename is wrong")
	}

	//Release fid beforedata
	err = v.status.freeSpace(file.Info.Offset-8, file.Info.Size+16)
	if err != nil {
		return err
	}

	err = v.index.Delete(fid)
	return err
}

func (v *Volume) NewFile(fid uint64, fileName string, size uint64) (f *File, err error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.index.Has(fid) {
		return nil, errors.New("fid already exists")
	}

	offset, err := v.newSpace(size + 16)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if e := v.status.freeSpace(offset, size+16); e != nil {
				panic(e)
			}
		}
	}()

	//write fid before and after content
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, fid)
	_, err = v.dataFile.WriteAt(b, int64(offset))
	if err != nil {
		return nil, err
	}
	_, err = v.dataFile.WriteAt(b, int64(offset+8+size))
	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Fid:      fid,
		Offset:   offset + 8,
		Size:     size,
		Ctime:    time.Now(),
		Mtime:    time.Now(),
		Atime:    time.Now(),
		FileName: fileName,
	}

	err = v.index.Set(fileInfo)
	if err != nil {
		return nil, err
	}
	return &File{DataFile: v.dataFile, Info: fileInfo}, nil

}

func (v *Volume) truncate() {
	currentDatafileSize := v.GetDatafileSize()
	if currentDatafileSize >= MaxVolumeSize {
		return
	}

	err := v.dataFile.Truncate(int64(currentDatafileSize) + int64(TruncateSize))
	if err != nil {
		panic(err)
	}

	err = v.status.freeSpace(currentDatafileSize, TruncateSize)
	if err != nil {
		panic(err)
	}
}

func (v *Volume) newSpace(size uint64) (uint64, error) {
	offset, err := v.status.newSpace(size)
	if err == nil {
		return offset, err
	}

	v.truncate()

	return v.status.newSpace(size)
}

func (v *Volume) Close() {
	v.mutex.Lock()
	v.status.spaceMutex.Lock()
	defer v.mutex.Unlock()

	v.dataFile.Close()
	v.dataFile = nil

	v.status.db.Close()
	v.status = nil

	v.index.Close()
	v.index = nil
}

func (v *Volume) GetDatafileSize() uint64 {
	fi, err := v.dataFile.Stat()
	if err != nil {
		panic(err)
	}
	return uint64(fi.Size())
}

func (v *Volume) GetMaxFreeSpace() uint64 {
	currentDatafileSize := v.GetDatafileSize()

	//have to get free space minSize(block)
	maxFreeSpace := v.status.getMaxFreeSpace()
	if currentDatafileSize < MaxVolumeSize {
		freeSpace := MaxVolumeSize - currentDatafileSize
		if freeSpace > maxFreeSpace {
			maxFreeSpace = freeSpace
		}
	}

	//8 * 2 bytes for fid
	if maxFreeSpace > 16 {
		return maxFreeSpace - 16
	}
	return 0

}
