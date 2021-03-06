package main

import (
	"fmt"
	"io"
	"os"

	lfs "github.com/bgould/go-littlefs"
)

var config = lfs.Config{
	ReadSize:      16,
	ProgSize:      16,
	BlockSize:     512,
	BlockCount:    32,
	CacheSize:     16,
	LookaheadSize: 16,
	BlockCycles:   500,
}

func main() {

	fmt.Printf("LittleFS Version: %08x\n", lfs.Version)

	// create/format/mount the filesystem
	fs := lfs.New(config, lfs.NewMemoryDevice(config))
	if err := fs.Format(); err != nil {
		fmt.Println("Could not format", err)
		os.Exit(1)
	}
	if err := fs.Mount(); err != nil {
		fmt.Println("Could not mount", err)
		os.Exit(1)
	}
	defer func() {
		if err := fs.Unmount(); err != nil {
			fmt.Println("Could not ummount", err)
			os.Exit(1)
		}
	}()

	// test an invalid operation to make sure it returns an appropriate error
	if err := fs.Rename("test.txt", "test2.txt"); err != nil {
		fmt.Println("Could not rename file (as expected):", err)
		os.Exit(1)
	}

	// try out some filesystem operations

	path := "/tmp"
	fmt.Println("making directory", path)
	if err := fs.Mkdir(path); err != nil {
		fmt.Println("Could not create "+path+" dir", err)
		os.Exit(1)
	}

	filepath := path + "/test.txt"
	fmt.Println("opening file")
	f, err := fs.OpenFile(filepath, lfs.O_WRONLY|lfs.O_CREAT)
	if err != nil {
		fmt.Println("Could not open file", err)
		os.Exit(1)
	}

	size, err := fs.Size()
	if err != nil {
		fmt.Println("Could not get filesystem size:", err.Error())
	} else {
		fmt.Println("Filesystem size:", size)
	}

	fmt.Println("truncating file")
	if err := f.Truncate(256); err != nil {
		fmt.Println("Could not trucate file", err)
		os.Exit(1)
	}

	for i := 0; i < 20; i++ {
		if _, err := f.Write([]byte("01234567890abcdef")); err != nil {
			fmt.Println("Could not write: %s", err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("closing file")
	if err := f.Close(); err != nil {
		fmt.Println("Could not close file", err)
		os.Exit(1)
	}

	if stat, err := fs.Stat(path); err != nil {
		fmt.Println("Could not stat dir", err)
		os.Exit(1)
	} else {
		fmt.Printf(
			"dir stat: name=%s size=%d dir=%t\n",
			stat.Name(), stat.Size(), stat.IsDir())
	}

	if stat, err := fs.Stat(filepath); err != nil {
		fmt.Println("Could not stat file", err)
		os.Exit(1)
	} else {
		fmt.Printf(
			"file stat: name=%s size=%d dir=%t\n",
			stat.Name(), stat.Size(), stat.IsDir())
	}

	fmt.Println("opening file read only")
	f, err = fs.OpenFile(filepath, lfs.O_RDONLY)
	if err != nil {
		fmt.Println("Could not open file", err)
		os.Exit(1)
	}
	defer f.Close()

	if size, err := f.Size(); err != nil {
		fmt.Printf("Failed getting file size: %v\n", err)
	} else {
		fmt.Printf("file size: %d\n", size)
	}

	buf := make([]byte, 57)
	for n := 0; n < 50; n++ {
		offset, err := f.Tell()
		if err != nil {
			fmt.Printf("Could not read offset with Tell: %s\n", err.Error())
		} else {
			fmt.Printf("reading from offset: %d\n", offset)
		}
		n, err := f.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("f.Read() error: %v\n", err.Error())
			}
			break
		}
		fmt.Printf("read %d bytes from file: `%s`", n, string(buf[:n]))
	}

	size, err = fs.Size()
	if err != nil {
		fmt.Println("Could not get filesystem size:", err.Error())
	} else {
		fmt.Println("Filesystem size:", size)
	}

	dir, err := fs.Open(path)
	if err != nil {
		fmt.Printf("Could not open directory %s: %v\n", path, err)
		os.Exit(1)
	}
	defer dir.Close()
	infos, err := dir.Readdir(0)
	_ = infos
	if err != nil {
		fmt.Printf("Could not read directory %s: %v\n", path, err)
		os.Exit(1)
	}
	for _, info := range infos {
		fmt.Printf("  directory entry: %s %d %t\n", info.Name(), info.Size(), info.IsDir())
	}
	fmt.Println("done")
	return
}
