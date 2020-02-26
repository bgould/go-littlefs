package main

import (
	"log"

	lfs "github.com/bgould/go-littlefs"
)

const debug = false

var config = lfs.Config{
	ReadSize:      16,
	ProgSize:      16,
	BlockSize:     512,
	BlockCount:    32,
	CacheSize:     16,
	LookaheadSize: 16,
	BlockCycles:   500,
}

var blankBlock = make([]byte, config.BlockSize)

func init() {
	for i := range blankBlock {
		blankBlock[i] = 0xff
	}
}

func main() {

	log.Printf("LittleFS Version: %08x\n", lfs.Version)

	// create/format/mount the filesystem
	fs := lfs.New(config, newMemoryDevice())
	if err := fs.Format(); err != nil {
		log.Fatalln("Could not format", err)
	}
	if err := fs.Mount(); err != nil {
		log.Fatalln("Could not mount", err)
	}
	defer func() {
		if err := fs.Unmount(); err != nil {
			log.Fatalln("Could not ummount", err)
		}
	}()

	// test an invalid operation to make sure it returns an appropriate error
	if err := fs.Rename("test.txt", "test2.txt"); err != nil {
		log.Println("Could not rename file (as expected):", err)
	}

	// try out some filesystem operations
	path := "/tmp"
	log.Println("making directory", path)
	if err := fs.Mkdir(path); err != nil {
		log.Fatalln("Could not create "+path+" dir", err)
	}
	f, err := fs.OpenFile("/tmp/test.txt", lfs.O_WRONLY|lfs.O_CREAT)
	if err != nil {
		log.Fatalln("Could not open file", err)
	}
	log.Println("truncating file")
	if err := f.Truncate(256); err != nil {
		log.Fatalln("Could not trucate file", err)
	}

}
