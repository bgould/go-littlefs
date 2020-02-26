package lfs

// #include <stdlib.h>
// #include "./go_lfs.h"
import "C"

import (
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

const (
	Version = C.LFS_VERSION

	FileTypeReg FileType = C.LFS_TYPE_REG
	FileTypeDir FileType = C.LFS_TYPE_DIR

	O_RDONLY OpenFlag = C.LFS_O_RDONLY // Open a file as read only
	O_WRONLY OpenFlag = C.LFS_O_WRONLY // Open a file as write only
	O_RDWR   OpenFlag = C.LFS_O_RDWR   // Open a file as read and write
	O_CREAT  OpenFlag = C.LFS_O_CREAT  // Create a file if it does not exist
	O_EXCL   OpenFlag = C.LFS_O_EXCL   // Fail if a file already exists
	O_TRUNC  OpenFlag = C.LFS_O_TRUNC  // Truncate the existing file to zero size
	O_APPEND OpenFlag = C.LFS_O_APPEND // Move to end of file on every write

	ErrOK                 = C.LFS_ERR_OK          // No error
	ErrIO           Error = C.LFS_ERR_IO          // Error during device operation
	ErrCorrupt      Error = C.LFS_ERR_CORRUPT     // Corrupted
	ErrNoEntry      Error = C.LFS_ERR_NOENT       // No directory entry
	ErrEntryExists  Error = C.LFS_ERR_EXIST       // Entry already exists
	ErrNotDir       Error = C.LFS_ERR_NOTDIR      // Entry is not a dir
	ErrIsDir        Error = C.LFS_ERR_ISDIR       // Entry is a dir
	ErrDirNotEmpty  Error = C.LFS_ERR_NOTEMPTY    // Dir is not empty
	ErrBadFileNum   Error = C.LFS_ERR_BADF        // Bad file number
	ErrFileTooLarge Error = C.LFS_ERR_FBIG        // File too large
	ErrInvalidParam Error = C.LFS_ERR_INVAL       // Invalid parameter
	ErrNoSpace      Error = C.LFS_ERR_NOSPC       // No space left on device
	ErrNoMemory     Error = C.LFS_ERR_NOMEM       // No more memory available
	ErrNoAttr       Error = C.LFS_ERR_NOATTR      // No data/attr available
	ErrNameTooLong  Error = C.LFS_ERR_NAMETOOLONG // File name too long
)

type OpenFlag int

type FileType uint

type Error int

func (err Error) Error() string {
	switch err {
	case ErrIO:
		return "littlefs: Error during device operation"
	case ErrCorrupt:
		return "littlefs: Corrupted"
	case ErrNoEntry:
		return "littlefs: No directory entry"
	case ErrEntryExists:
		return "littlefs: Entry already exists"
	case ErrNotDir:
		return "littlefs: Entry is not a dir"
	case ErrIsDir:
		return "littlefs: Entry is a dir"
	case ErrDirNotEmpty:
		return "littlefs: Dir is not empty"
	case ErrBadFileNum:
		return "littlefs: Bad file number"
	case ErrFileTooLarge:
		return "littlefs: File too large"
	case ErrInvalidParam:
		return "littlefs: Invalid parameter"
	case ErrNoSpace:
		return "littlefs: No space left on device"
	case ErrNoMemory:
		return "littlefs: No more memory available"
	case ErrNoAttr:
		return "littlefs: No data/attr available"
	case ErrNameTooLong:
		return "littlefs: File name too long"
	default:
		return "littlefs: Unknown error"
	}
}

type Config struct {
	ReadSize      uint32
	ProgSize      uint32
	BlockSize     uint32
	BlockCount    uint32
	CacheSize     uint32
	LookaheadSize uint32
	BlockCycles   int32
}

type LFS struct {
	ptr unsafe.Pointer
	lfs *C.struct_lfs
	//cfg C.struct_lfs_config
	//cfg C.struct_lfs_config
	cfg *C.struct_lfs_config
}

func New(config Config, blockdev BlockDevice) *LFS {
	ptr1 := C.go_lfs_new_lfs()
	ptr2 := C.go_lfs_new_lfs_config()
	lfs := &LFS{
		//lfs: (*C.struct_lfs)(C.malloc(C.sizeof_struct_lfs)),
		//cfg: (*C.struct_lfs_config)(C.malloc(C.sizeof_struct_lfs_config)),
		lfs: ptr1,
		cfg: ptr2,
	}
	lfs.ptr = gopointer.Save(lfs) // save this to prevent GC until Close() is called?
	//lfs.cfg = C.go_lfs_set_callbacks(
	*lfs.cfg = C.struct_lfs_config{
		context:        gopointer.Save(blockdev),
		read_size:      C.lfs_size_t(config.ReadSize),
		prog_size:      C.lfs_size_t(config.ProgSize),
		block_size:     C.lfs_size_t(config.BlockSize),
		block_count:    C.lfs_size_t(config.BlockCount),
		cache_size:     C.lfs_size_t(config.CacheSize),
		lookahead_size: C.lfs_size_t(config.LookaheadSize),
		block_cycles:   C.int32_t(config.BlockCycles),
	}
	C.go_lfs_set_callbacks(lfs.cfg)
	//)
	return lfs
}

func (l *LFS) Mount() error {
	return errval(C.lfs_mount(l.lfs, l.cfg))
}

func (l *LFS) Format() error {
	return errval(C.lfs_format(l.lfs, l.cfg))
}

func (l *LFS) Unmount() error {
	return errval(C.lfs_unmount(l.lfs))
	return nil
}

func (l *LFS) Mkdir(path string) error {
	cs := cstring(path)
	defer C.free(unsafe.Pointer(cs))
	return errval(C.lfs_mkdir(l.lfs, cs))
}

func (l *LFS) Remove(path string) error {
	cs := cstring(path)
	defer C.free(unsafe.Pointer(cs))
	return errval(C.lfs_remove(l.lfs, cs))
}

func (l *LFS) Rename(oldPath string, newPath string) error {
	cs1, cs2 := cstring(oldPath), cstring(newPath)
	defer C.free(unsafe.Pointer(cs1))
	defer C.free(unsafe.Pointer(cs2))
	return errval(C.lfs_rename(l.lfs, cs1, cs2))
}

func (l *LFS) OpenFile(path string, flags OpenFlag) (*File, error) {
	cs := cstring(path)
	defer C.free(unsafe.Pointer(cs))
	file := &File{lfs: l}
	errno := C.lfs_file_open(l.lfs, &file.fptr, cs, C.int(flags))
	err := errval(errno)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type File struct {
	lfs  *LFS
	fptr C.struct_lfs_file
}

// Synchronize a file on storage
//
// Any pending writes are written out to storage.
// Returns a negative error code on failure.
func (f *File) Sync() error {
	return errval(C.lfs_file_sync(f.lfs.lfs, &f.fptr))
}

// Close a file
//
// Any pending writes are written out to storage as though
// sync had been called and releases any allocated resources.
//
// Returns a negative error code on failure.
func (f *File) Close() error {
	return errval(C.lfs_file_close(f.lfs.lfs, &f.fptr))
}

// Truncates the size of the file to the specified size
//
// Returns a negative error code on failure.
func (f *File) Truncate(size uint32) error {
	return errval(C.lfs_file_truncate(f.lfs.lfs, &f.fptr, C.lfs_off_t(size)))
}

// would be nice to use C.CString instead, but TinyGo doesn't seem to support
func cstring(s string) *C.char {
	ptr := C.malloc(C.ulong(len(s) + 1))
	buf := (*[1 << 28]byte)(ptr)[: len(s)+1 : len(s)+1]
	copy(buf, s)
	buf[len(s)] = 0
	return (*C.char)(ptr)
}

func errval(errno C.int) error {
	if errno < ErrOK {
		return Error(errno)
	}
	return nil
}

type BlockDevice interface {
	ReadBlock(block uint32, offset uint32, buf []byte) error
	ProgramBlock(block uint32, offset uint32, buf []byte) error
	EraseBlock(block uint32) error
	Sync() error
}
