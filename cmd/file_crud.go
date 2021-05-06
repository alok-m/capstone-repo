package cmd

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"strconv"

	"github.com/alok-m/capstone-codebase/storage"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fileCrud)
}

var fileCrud = &cobra.Command{
	Use:   "file",
	Short: "crud for file",
	Long:  `file crud (creating and deleteing)`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("file crud main")
		// write fileid volumedir input
		if args[0] == "write" {
			// data := make([]byte, 1024)
			// rand.Read(data)
			data, err := retrieveROM(args[3])
			if err != nil {
				log.Fatal(err)
			}
			v, err := storage.NewVolume(args[2], 0)
			if err != nil {
				log.Fatal(err)
			}
			defer v.Close()
			a, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			file, err := v.NewFile(a, args[1], uint64(len(data)))
			if err != nil {
				log.Fatal(err)
			}
			n, err := file.Write(data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(strconv.FormatInt(int64(n), 10) + " bytes written with fid " + strconv.FormatUint(file.Info.Fid, 10))
			fmt.Printf("crc32 checksum:%d\n ", crc32.ChecksumIEEE(data))
		}
		// read fileID volumedir outupt
		if args[0] == "read" {
			v, err := storage.NewVolume(args[2], 0)
			if err != nil {
				log.Fatal(err)
			}
			id, err := strconv.ParseUint(args[1], 10, 64)
			fmt.Printf("ID to be read :%d\n ", id)
			if err != nil {
				log.Fatal(err)
			}
			file, err := v.Get(id)
			if err != nil {
				log.Fatal(err)
			}
			data := make([]byte, file.Info.Size)
			file.Read(data)
			f, err := os.Create(args[3])
			if err != nil {
				log.Fatal(err)
			}
			n, err := f.Write(data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(n, " bytes written to output")
			fmt.Printf("crc32 checksum:%d\n ", crc32.ChecksumIEEE(data))

		}

		if args[0] == "delete" {
			v, err := storage.NewVolume(args[2], 0)
			if err != nil {
				log.Fatal(err)
			}
			id, err := strconv.ParseUint(args[1], 10, 64)
			fmt.Println("ID to be read : ", id)
			if err != nil {
				log.Fatal(err)
			}
			err = v.Delete(id, args[1])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("File with fid %d has been deleted\n", id)
		}
	},
}

func retrieveROM(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}
