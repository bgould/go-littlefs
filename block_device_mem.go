package lfs

import (
	"log"
)

// MemBlockDevice is a block device implementation backed by a byte slice
type MemBlockDevice struct {
	config     Config
	memory     []byte
	blankBlock []byte
}

func NewMemoryDevice(config Config) *MemBlockDevice {
	dev := &MemBlockDevice{
		config:     config,
		memory:     make([]byte, config.BlockSize*config.BlockCount),
		blankBlock: make([]byte, config.BlockSize),
	}
	for i := range dev.blankBlock {
		dev.blankBlock[i] = 0xff
	}
	for i := uint32(0); i < config.BlockCount; i++ {
		if err := dev.EraseBlock(i); err != nil {
			log.Fatalf("could not initialize block %d: %s", i, err.Error())
		}
	}
	return dev
}

func (bd *MemBlockDevice) ReadBlock(block uint32, offset uint32, buf []byte) error {
	copy(buf, bd.memory[bd.config.BlockSize*block+offset:])
	return nil
}

func (bd *MemBlockDevice) ProgramBlock(block uint32, offset uint32, buf []byte) error {
	copy(bd.memory[bd.config.BlockSize*block+offset:], buf)
	return nil
}

func (bd *MemBlockDevice) EraseBlock(block uint32) error {
	copy(bd.memory[bd.config.BlockSize*block:], bd.blankBlock)
	return nil
}

func (bd *MemBlockDevice) Sync() error {
	return nil
}
