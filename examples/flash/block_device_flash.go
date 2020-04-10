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
		BlockSize:     flash.SectorSize,
		BlockCount:    0,
		CacheSize:     flash.PageSize,
		LookaheadSize: flash.PageSize,
		BlockCycles:   100,
	}
}

func NewFlashBlockDevice(config lfs.Config, device *flash.Device, startBlock uint32) *FlashBlockDevice {
	return &FlashBlockDevice{cfg: config, dev: device, startBlock: startBlock}
}

// FlashBlockDevice is a LittleFS block device implementation that is backed
// by a SPI NOR flash chip
type FlashBlockDevice struct {
	cfg        lfs.Config
	dev        *flash.Device
	startBlock uint32
}

func (bd *FlashBlockDevice) ReadBlock(block uint32, offset uint32, buf []byte) error {
	addr := bd.cfg.BlockSize*(bd.startBlock+block) + offset
	if debug {
		fmt.Printf("lfs: ReadBlock(): %v, %v, %v, %v, %08x\n", block, offset, len(buf), bd.startBlock, addr)
	}
	_, err := bd.dev.ReadAt(buf, int64(addr))
	return err
}

func (bd *FlashBlockDevice) ProgramBlock(block uint32, offset uint32, buf []byte) error {
	addr := bd.cfg.BlockSize*(bd.startBlock+block) + offset
	if debug {
		fmt.Printf("lfs: ProgramBlock(): %v, %v, %v, %v, %08x\n", block, offset, len(buf), bd.startBlock, addr)
	}
	_, err := bd.dev.WriteAt(buf, int64(addr))
	return err
}

func (bd *FlashBlockDevice) EraseBlock(block uint32) error {
	eraseBlock := bd.startBlock + block
	if debug {
		fmt.Printf("lfs: EraseBlock(): %v %v %v\n", block, bd.startBlock, eraseBlock)
	}
	err := bd.dev.EraseSector(eraseBlock)
	return err
}

func (bd *FlashBlockDevice) Sync() error {
	if debug {
		fmt.Printf("lfs: Sync()\n")
	}
	return nil
}
