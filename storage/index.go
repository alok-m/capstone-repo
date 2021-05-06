package storage

import (
	"io"
)

//Index interface
type Index interface {
	Has(fid uint64) bool
	Get(fid uint64) (*FileInfo, error)
	Set(info *FileInfo) error
	Delete(fid uint64) error
	io.Closer
}
