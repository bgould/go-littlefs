package lfs

import (
	"fmt"
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

import "C"

const (
	debug bool = false
)

type BlockDevice interface {
	ReadBlock(block uint32, offset uint32, buf []byte) error
	ProgramBlock(block uint32, offset uint32, buf []byte) error
	EraseBlock(block uint32) error
	Sync() error
}

//export go_lfs_block_device_read
func go_lfs_block_device_read(ctx unsafe.Pointer, block uint32, offset uint32, buf unsafe.Pointer, size int) int {
	if debug {
		fmt.Printf("go_lfs_block_device_read: %v, %v, %v, %v, %v\n", ctx, block, offset, buf, size)
	}
	buffer := (*[1 << 28]byte)(buf)[:size:size]
	if err := restore(ctx).ReadBlock(block, offset, buffer); err != nil {
		if debug {
			println("read error:", err)
		}
		return int(ErrIO)
	}
	return ErrOK
}

//export go_lfs_block_device_prog
func go_lfs_block_device_prog(ctx unsafe.Pointer, block uint32, offset uint32, buf unsafe.Pointer, size int) int {
	if debug {
		fmt.Printf("go_lfs_block_device_prog: %v, %v, %v, %v, %v\n", ctx, block, offset, buf, size)
	}
	buffer := (*[1 << 28]byte)(buf)[:size:size]
	if err := restore(ctx).ProgramBlock(block, offset, buffer); err != nil {
		if debug {
			println("program error:", err)
		}
		return int(ErrIO)
	}
	return ErrOK
}

//export go_lfs_block_device_erase
func go_lfs_block_device_erase(ctx unsafe.Pointer, block uint32) int {
	if debug {
		fmt.Printf("go_lfs_block_device_erase: %v, %v\n", ctx, block)
	}
	if err := restore(ctx).EraseBlock(block); err != nil {
		if debug {
			println("erase error:", err)
		}
		return int(ErrIO)
	}
	return ErrOK
}

//export go_lfs_block_device_sync
func go_lfs_block_device_sync(ctx unsafe.Pointer) int {
	if debug {
		fmt.Printf("go_lfs_block_device_sync: %v\n", ctx)
	}
	if err := restore(ctx).Sync(); err != nil {
		if debug {
			println("sync error:", err)
		}
		return int(ErrIO)
	}
	return ErrOK
}

func restore(ptr unsafe.Pointer) BlockDevice {
	return gopointer.Restore(ptr).(BlockDevice)
}
