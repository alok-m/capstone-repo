package storage

import (
	"crypto/rand"
	"crypto/sha1"
	"io/ioutil"
	"os"
	"testing"
)

func TestVolumeAndFile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test_volume")
	defer os.RemoveAll(dir)

	v, err := NewVolume(dir, 0)
	if err != nil {
		t.Error(err)
	}
	// crud sizes
	for i, size := range []uint64{1, 100, 1024, 1024 * 1024, 1024 * 1024 * 10} {
		file, err := v.NewFile(uint64(i), "test_file.1", size)
		if err != nil {
			t.Error(err)
		}
		// write test
		data := make([]byte, size)
		rand.Read(data)
		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
		}
		// read test
		file2, err := v.Get(file.Info.Fid)
		if err != nil {
			t.Error(err)
		}

		data2 := make([]byte, size)
		file2.Read(data2)

		if sha1.Sum(data) != sha1.Sum(data2) {
			t.Error("data wrong")
		}
		// delete test
		err = v.Delete(file.Info.Fid, "test_file.1")
		if err != nil {
			t.Error(err)
		}

		file3, err := v.Get(file.Info.Fid)
		if err == nil || file3 != nil {
			t.Error("delete failed?")
		}
	}

	// for _, nFiles := range []int{10, 100, 100 * 100} {
	// 	for j := 0; j < nFiles; j++ {
	// 		size := r.Intn(100-1) + 1
	// 		file, err := v.NewFile(uint64(j), "test_file."+fmt.Sprint(j), uint64(size))
	// 		if err != nil {
	// 			t.Errorf("could not create file")
	// 		}

	// 		data := make([]byte, size)
	// 		rand.Read(data)
	// 		_, err = file.Write(data)
	// 		if err != nil {
	// 			t.Error(err)
	// 		}

	// 		// 	}

	// 		// 	for j := 1; j < nFiles; j++ {
	// 		// 		id := r.Intn(j)
	// 		// 		file, err := v.Get(uint64(id))
	// 		// 		if err != nil {
	// 		// 			t.Errorf("fid not found")
	// 		// 		}
	// 		// 		data := make([]byte, file.Info.Size)
	// 		// 		file.Read(data)
	// 		// 	}
	// 		// }
	// 	}
	// }
}
