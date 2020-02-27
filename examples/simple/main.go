package main

import (
	"io"
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

	filepath := path + "/test.txt"
	log.Println("opening file")
	f, err := fs.OpenFile(filepath, lfs.O_WRONLY|lfs.O_CREAT)
	if err != nil {
		log.Fatalln("Could not open file", err)
	}

	size, err := fs.Size()
	if err != nil {
		log.Println("Could not get filesystem size:", err.Error())
	} else {
		log.Println("Filesystem size:", size)
	}

	log.Println("truncating file")
	if err := f.Truncate(256); err != nil {
		log.Fatalln("Could not trucate file", err)
	}

	for i := 0; i < 20; i++ {
		if _, err := f.Write([]byte("01234567890abcdef")); err != nil {
			log.Fatalln("Could not write: %s", err.Error())
		}
	}

	log.Println("closing file")
	if err := f.Close(); err != nil {
		log.Fatalln("Could not close file", err)
	}

	if stat, err := fs.Stat(path); err != nil {
		log.Fatalln("Could not stat dir", err)
	} else {
		log.Printf(
			"dir stat: name=%s size=%d dir=%t\n",
			stat.Name(), stat.Size(), stat.IsDir())
	}

	if stat, err := fs.Stat(filepath); err != nil {
		log.Fatalln("Could not stat file", err)
	} else {
		log.Printf(
			"file stat: name=%s size=%d dir=%t\n",
			stat.Name(), stat.Size(), stat.IsDir())
	}

	log.Println("opening file read only")
	f, err = fs.OpenFile(filepath, lfs.O_RDONLY)
	if err != nil {
		log.Fatalln("Could not open file", err)
	}
	defer f.Close()

	buf := make([]byte, 57)
	for n := 0; n < 50; n++ {
		offset, err := f.Tell()
		if err != nil {
			log.Printf("Could not read offset with Tell: %s\n", err.Error())
		} else {
			log.Printf("reading from offset: %d\n", offset)
		}
		n, err := f.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("f.Read() error: %v\n", err.Error())
			}
			break
		}
		log.Printf("read %d bytes from file: `%s`", n, string(buf[:n]))
	}

	size, err = fs.Size()
	if err != nil {
		log.Println("Could not get filesystem size:", err.Error())
	} else {
		log.Println("Filesystem size:", size)
	}
}
