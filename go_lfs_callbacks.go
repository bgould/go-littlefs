package lfs

import "C"
import (
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

type BlockDevice interface {
	ReadBlock(block uint32, offset uint32, buf []byte) error
	ProgramBlock(block uint32, offset uint32, buf []byte) error
	EraseBlock(block uint32) error
	Sync() error
}

//export go_lfs_block_device_read
func go_lfs_block_device_read(ctx unsafe.Pointer, block uint32, offset uint32, buf unsafe.Pointer, size int) int {
	buffer := (*[1 << 28]byte)(buf)[:size:size]
	restore(ctx).ReadBlock(block, offset, buffer)
	return ErrOK
}

//export go_lfs_block_device_prog
func go_lfs_block_device_prog(ctx unsafe.Pointer, block uint32, offset uint32, buf unsafe.Pointer, size int) int {
	buffer := (*[1 << 28]byte)(buf)[:size:size]
	restore(ctx).ProgramBlock(block, offset, buffer)
	return ErrOK
}

//export go_lfs_block_device_erase
func go_lfs_block_device_erase(ctx unsafe.Pointer, block uint32) int {
	restore(ctx).EraseBlock(block)
	return ErrOK
}

//export go_lfs_block_device_sync
func go_lfs_block_device_sync(ctx unsafe.Pointer) int {
	restore(ctx).Sync()
	return ErrOK
}

func restore(ptr unsafe.Pointer) BlockDevice {
	return gopointer.Restore(ptr).(BlockDevice)
}
