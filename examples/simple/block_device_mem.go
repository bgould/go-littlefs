package main

import (
	"fmt"
	"log"
)

func newMemoryDevice() *MemBlockDevice {
	dev := &MemBlockDevice{
		memory: make([]byte, config.BlockSize*config.BlockCount),
	}
	for i := uint32(0); i < config.BlockCount; i++ {
		if err := dev.EraseBlock(i); err != nil {
			log.Fatalf("could not initialize block %d: %s", i, err.Error())
		}
	}
	return dev
}

// MemBlockDevice is a block device implementation backed by a byte slice
type MemBlockDevice struct {
	memory []byte
}

func (bd *MemBlockDevice) ReadBlock(block uint32, offset uint32, buf []byte) error {
	if debug {
		fmt.Printf("lfs: ReadBlock(): %v, %v, %v\n", block, offset, len(buf))
	}
	copy(buf, bd.memory[config.BlockSize*block+offset:])
	return nil
}

func (bd *MemBlockDevice) ProgramBlock(block uint32, offset uint32, buf []byte) error {
	if debug {
		fmt.Printf("lfs: ProgramBlock(): %v, %v, %v\n", block, offset, len(buf))
	}
	copy(bd.memory[config.BlockSize*block+offset:], buf)
	return nil
}

func (bd *MemBlockDevice) EraseBlock(block uint32) error {
	if debug {
		fmt.Printf("lfs: EraseBlock(): %v\n", block)
	}
	copy(bd.memory[config.BlockSize*block:], blankBlock)
	return nil
}

func (bd *MemBlockDevice) Sync() error {
	if debug {
		fmt.Printf("lfs: Sync()\n")
	}
	return nil
}
