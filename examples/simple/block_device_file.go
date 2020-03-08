// +build !tinygo

package main

import (
	"fmt"
	"log"
	"os"
)

const debug = false

var blankBlock = make([]byte, config.BlockSize)

func init() {
	for i := range blankBlock {
		blankBlock[i] = 0xff
	}
}

func newFileDevice(path string) *FileBlockDevice {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal("Could not create temporary file: " + err.Error())
	}
	for i := uint32(0); i < config.BlockCount; i++ {
		if _, err := file.Write(blankBlock); err != nil {
			log.Fatalf("could not write block %d: %s", i, err.Error())
		}
	}
	return &FileBlockDevice{File: file}
}

// FileBlockDevice is a LittleFS block device implementation that is backed
// by a single file on the filesystem of the host OS
type FileBlockDevice struct {
	File *os.File
}

func (bd *FileBlockDevice) ReadBlock(block uint32, offset uint32, buf []byte) error {
	if debug {
		fmt.Printf("lfs: ReadBlock(): %v, %v, %v\n", block, offset, len(buf))
	}
	_, err := bd.File.ReadAt(buf, int64(config.BlockSize*block+offset))
	return err
}

func (bd *FileBlockDevice) ProgramBlock(block uint32, offset uint32, buf []byte) error {
	if debug {
		fmt.Printf("lfs: ProgramBlock(): %v, %v, %v\n", block, offset, len(buf))
	}
	_, err := bd.File.WriteAt(buf, int64(config.BlockSize*block+offset))
	return err
}

func (bd *FileBlockDevice) EraseBlock(block uint32) error {
	if debug {
		fmt.Printf("lfs: EraseBlock(): %v\n", block)
	}
	_, err := bd.File.WriteAt(blankBlock, int64(config.BlockSize*block))
	return err
}

func (bd *FileBlockDevice) Sync() error {
	if debug {
		fmt.Printf("lfs: Sync()\n")
	}
	return nil
}
