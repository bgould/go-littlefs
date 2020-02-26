package lfs

import "C"
import (
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

//export go_lfs_block_device_read
func go_lfs_block_device_read(ctx unsafe.Pointer, block uint32, offset uint32, buf unsafe.Pointer, size int) int {
	//fmt.Printf("go_lfs_block_device_read: %v, %v, %v, %v, %v\n", ctx, block, offset, buf, size)
	buffer := (*[1 << 28]byte)(buf)[:size:size]
	restore(ctx).ReadBlock(block, offset, buffer)
	return ErrOK
}

//export go_lfs_block_device_prog
func go_lfs_block_device_prog(ctx unsafe.Pointer, block uint32, offset uint32, buf unsafe.Pointer, size int) int {
	//fmt.Printf("go_lfs_block_device_prog: %v, %v, %v, %v, %v\n", ctx, block, offset, buf, size)
	buffer := (*[1 << 28]byte)(buf)[:size:size]
	restore(ctx).ProgramBlock(block, offset, buffer)
	return ErrOK
}

//export go_lfs_block_device_erase
func go_lfs_block_device_erase(ctx unsafe.Pointer, block uint32) int {
	//fmt.Printf("go_lfs_block_device_erase: %v, %v\n", ctx, block)
	restore(ctx).EraseBlock(block)
	return ErrOK
}

//export go_lfs_block_device_sync
func go_lfs_block_device_sync(ctx unsafe.Pointer) int {
	//fmt.Printf("go_lfs_block_device_sync: %v\n", ctx)
	restore(ctx).Sync()
	return ErrOK
}

func restore(ptr unsafe.Pointer) BlockDevice {
	return gopointer.Restore(ptr).(BlockDevice)
}
