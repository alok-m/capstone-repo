package storage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	ReversedsizeOffsetPrefix = '\x11' //key= "\x01"+Reversesize(8 byte)+offset(8 byte) value=[]
	OffsetSizePrefix         = '\x22' //key= "\x02"+offset(8 byte)+size(8 byte) value=[]
)

// Status stores location of free blocks
type Status struct {
	path string
	db   *leveldb.DB

	spaceMutex sync.Mutex
}

//Newstatus based on directory and volume id
func NewStatus(dir string, vid uint64) (status *Status, err error) {
	path := filepath.Join(dir, strconv.FormatUint(vid, 10)+".status")
	status = new(Status)
	status.path = path
	status.db, err = leveldb.OpenFile(path, nil)
	return status, err
}

func (s *Status) newSpace(size uint64) (offset uint64, err error) {
	s.spaceMutex.Lock()
	defer s.spaceMutex.Unlock()
	//Here is stored in reverse order according to size, so that the largest space is obtained first
	iter := s.db.NewIterator(util.BytesPrefix([]byte{ReversedsizeOffsetPrefix}), nil)
	defer iter.Release()

	iter.Next()
	key := iter.Key()
	if key == nil {
		return 0, errors.New("can't get free space")
	}
	freeSize := binary.BigEndian.Uint64(key[1:9]) ^ (^uint64(0))
	if freeSize < size {
		return 0, errors.New("can't get free space")
	}

	offset = binary.BigEndian.Uint64(key[9:])
	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return 0, err
	}
	transaction.Delete(key, nil)
	key = getOffsetSizeKey(offset, freeSize)
	transaction.Delete(key, nil)

	if freeSize > size {
		key = getReversedsizeOffset(offset+size, freeSize-size)
		transaction.Put(key, nil, nil)

		key = getOffsetSizeKey(offset+size, freeSize-size)
		transaction.Put(key, nil, nil)
	}

	err = transaction.Commit()
	return offset, err
}

func (s *Status) freeSpace(offset uint64, size uint64) error {
	s.spaceMutex.Lock()
	defer s.spaceMutex.Unlock()

	iter := s.db.NewIterator(util.BytesPrefix([]byte{OffsetSizePrefix}), nil)
	defer iter.Release()

	key := getOffsetSizeKey(offset, 0)
	iter.Seek(key)

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return err
	}

	//find adjacent to free block
	key = iter.Key()
	if len(key) != 0 {
		nOffset := binary.BigEndian.Uint64(key[1:9])
		nSize := binary.BigEndian.Uint64(key[9:])
		if nOffset < offset+size {
			panic(fmt.Errorf("nOffset: %d < offset: %d + size: %d", nOffset, offset, size))
			//if nOffset == offset {
			//	transaction.Discard()
			//return errors.New("space already free")
		} else if nOffset == offset+size {
			transaction.Delete(key, nil)
			size += nSize

			key = getReversedsizeOffset(nOffset, nSize)
			transaction.Delete(key, nil)
		}
	}

	//find adjacent to free block , create new block
	iter.Prev()
	key = iter.Key()
	if len(key) != 0 {
		pOffset := binary.BigEndian.Uint64(key[1:9])
		pSize := binary.BigEndian.Uint64(key[9:])
		if pOffset+pSize > offset {
			panic(fmt.Errorf("pOffset: %d + pSize: %d > offset: %d", pOffset, pSize, offset))
			//transaction.Discard()
			//return errors.New("space alread free")
		} else if pOffset+pSize == offset {
			transaction.Delete(key, nil)
			offset = pOffset
			size += pSize

			key = getReversedsizeOffset(pOffset, pSize)
			transaction.Delete(key, nil)
		}
	}

	key = getOffsetSizeKey(offset, size)
	transaction.Put(key, nil, nil)

	key = getReversedsizeOffset(offset, size)
	transaction.Put(key, nil, nil)

	return transaction.Commit()
}

func (s *Status) getMaxFreeSpace() uint64 {
	iter := s.db.NewIterator(util.BytesPrefix([]byte{ReversedsizeOffsetPrefix}), nil)
	defer iter.Release()

	iter.Next()
	key := iter.Key()
	if len(key) == 0 {
		return 0
	}

	freeSize := binary.BigEndian.Uint64(key[1:9]) ^ (^uint64(0))
	return freeSize
}

func getOffsetSizeKey(offset, size uint64) []byte {
	key := make([]byte, 1+16)
	key[0] = OffsetSizePrefix
	binary.BigEndian.PutUint64(key[1:9], offset)
	binary.BigEndian.PutUint64(key[9:], size)
	return key
}

func getReversedsizeOffset(offset, size uint64) []byte {
	key := make([]byte, 1+16)
	key[0] = ReversedsizeOffsetPrefix
	binary.BigEndian.PutUint64(key[9:], offset)
	binary.BigEndian.PutUint64(key[1:9], size^(^uint64(0)))
	return key
}
