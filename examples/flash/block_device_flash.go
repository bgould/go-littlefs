package flash_example

import (
	"fmt"

	lfs "github.com/bgould/go-littlefs"

	"tinygo.org/x/drivers/flash"
)

func FlashLFSConfig() lfs.Config {
	return lfs.Config{
		ReadSize:      flash.PageSize,
		ProgSize:      flash.PageSize,
		BlockSize:     flash.BlockSize,
		BlockCount:    0,
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	}
}

func NewFlashBlockDevice(config lfs.Config, device *flash.Device) *FlashBlockDevice {
	return &FlashBlockDevice{cfg: config, dev: device}
}

// FlashBlockDevice is a LittleFS block device implementation that is backed
// by a SPI NOR flash chip
type FlashBlockDevice struct {
	cfg lfs.Config
	dev *flash.Device
}

func (bd *FlashBlockDevice) ReadBlock(block uint32, offset uint32, buf []byte) error {
	if debug {
		fmt.Printf("lfs: ReadBlock(): %v, %v, %v\n", block, offset, len(buf))
	}
	err := bd.dev.ReadBuffer(bd.cfg.BlockSize*block+offset, buf)
	return err
}

func (bd *FlashBlockDevice) ProgramBlock(block uint32, offset uint32, buf []byte) error {
	if debug {
		fmt.Printf("lfs: ProgramBlock(): %v, %v, %v\n", block, offset, len(buf))
	}
	_, err := bd.dev.WriteBuffer(bd.cfg.BlockSize*block+offset, buf)
	return err
}

func (bd *FlashBlockDevice) EraseBlock(block uint32) error {
	if debug {
		fmt.Printf("lfs: EraseBlock(): %v\n", block)
	}
	bd.dev.EraseBlock(block)
	err := bd.dev.EraseBlock(block)
	return err
}

func (bd *FlashBlockDevice) Sync() error {
	if debug {
		fmt.Printf("lfs: Sync()\n")
	}
	return nil
}
