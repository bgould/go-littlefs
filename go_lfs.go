package lfs

// #include <string.h>
// #include <stdlib.h>
// #include "./go_lfs.h"
import "C"

import (
	"errors"
	"io"
	"os"
	"time"
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
	cfg *C.struct_lfs_config
}

type Info struct {
	fileType FileType
	size     uint32
	name     string
}

var _ os.FileInfo = (*Info)(nil)

func (info *Info) Name() string {
	return info.name
}

func (info *Info) Size() int64 {
	return int64(info.size)
}

func (info *Info) IsDir() bool {
	return info.fileType == FileTypeDir
}

func (info *Info) Sys() interface{} {
	return nil
}

func (info *Info) Mode() os.FileMode {
	v := os.FileMode(0777)
	if info.IsDir() {
		v |= os.ModeDir
	}
	return v
}

func (info *Info) ModTime() time.Time {
	return time.Time{}
}

func New(config Config, blockdev BlockDevice) *LFS {
	lfs := &LFS{
		// sizeof_* doesn't seem to work in TinyGo, so using C helper funcs instead
		//lfs: (*C.struct_lfs)(C.malloc(C.sizeof_struct_lfs)),
		//cfg: (*C.struct_lfs_config)(C.malloc(C.sizeof_struct_lfs_config)),
		lfs: C.go_lfs_new_lfs(),
		cfg: C.go_lfs_new_lfs_config(),
	}
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
	lfs.ptr = gopointer.Save(lfs) // save this to prevent GC until Close() is called?
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

func (l *LFS) Stat(path string) (*Info, error) {
	cs := cstring(path)
	defer C.free(unsafe.Pointer(cs))
	info := C.struct_lfs_info{}
	if err := errval(C.lfs_stat(l.lfs, cs, &info)); err != nil {
		return nil, err
	}
	return &Info{
		fileType: FileType(info._type),
		size:     uint32(info.size),
		name:     gostring(&info.name[0]),
	}, nil
}

func (l *LFS) Mkdir(path string) error {
	cs := cstring(path)
	defer C.free(unsafe.Pointer(cs))
	return errval(C.lfs_mkdir(l.lfs, cs))
}

func (l *LFS) Open(path string) (*File, error) {
	return l.OpenFile(path, O_RDONLY)
}

func (l *LFS) OpenFile(path string, flags OpenFlag) (*File, error) {

	cs := cstring(path)
	defer C.free(unsafe.Pointer(cs))
	file := &File{lfs: l, name: path}

	var fileType FileType
	info := C.struct_lfs_info{}
	if err := errval(C.lfs_stat(l.lfs, cs, &info)); err == nil {
		fileType = FileType(info._type)
	}

	var errno C.int
	if fileType == FileTypeDir {
		file.typ = FileTypeDir
		file.hndl = unsafe.Pointer(C.go_lfs_new_lfs_dir())
		errno = C.lfs_dir_open(l.lfs, file.dirptr(), cs)
	} else {
		file.typ = FileTypeReg
		file.hndl = unsafe.Pointer(C.go_lfs_new_lfs_file())
		errno = C.lfs_file_open(l.lfs, file.fileptr(), cs, C.int(flags))
	}

	if err := errval(errno); err != nil {
		if file.hndl != nil {
			C.free(file.hndl)
			file.hndl = nil
		}
		return nil, err
	}

	return file, nil
}

// Finds the current size of the filesystem
//
// Note: Result is best effort. If files share COW structures, the returned
// size may be larger than the filesystem actually is.
//
// Returns the number of allocated blocks, or a negative error code on failure.
func (l *LFS) Size() (n int, err error) {
	errno := C.int(C.lfs_fs_size(l.lfs))
	if errno < 0 {
		return 0, errval(errno)
	}
	return int(errno), nil
}

type File struct {
	lfs  *LFS
	typ  FileType
	hndl unsafe.Pointer
	//fptr C.struct_lfs_file
	//dptr *C.struct_lfs_dir
	name string
}

func (f *File) dirptr() *C.struct_lfs_dir {
	return (*C.struct_lfs_dir)(f.hndl)
}

func (f *File) fileptr() *C.struct_lfs_file {
	return (*C.struct_lfs_file)(f.hndl)
}

// Name returns the name of the file as presented to OpenFile
func (f *File) Name() string {
	return f.name
}

// Close the file; any pending writes are written out to storage
func (f *File) Close() error {
	if f.hndl != nil {
		defer func() {
			C.free(f.hndl)
			f.hndl = nil
		}()
		switch f.typ {
		case FileTypeReg:
			return errval(C.lfs_file_close(f.lfs.lfs, f.fileptr()))
		case FileTypeDir:
			return errval(C.lfs_dir_close(f.lfs.lfs, f.dirptr()))
		default:
			panic("lfs: unknown typ for file handle")
		}
	}
	return nil
}

func (f *File) Read(buf []byte) (n int, err error) {
	if f.IsDir() {
		return 0, ErrIsDir
	}
	bufptr := unsafe.Pointer(&buf[0])
	buflen := C.lfs_size_t(len(buf))
	errno := C.int(C.lfs_file_read(f.lfs.lfs, f.fileptr(), bufptr, buflen))
	if errno > 0 {
		return int(errno), nil
	} else if errno == 0 {
		// TODO: any extra checks needed here?
		return 0, io.EOF
	} else {
		return 0, errval(errno)
	}
}

// Seek changes the position of the file
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	errno := C.int(C.lfs_file_seek(f.lfs.lfs, f.fileptr(), C.lfs_soff_t(offset), C.int(whence)))
	if errno < 0 {
		return -1, errval(errno)
	}
	return int64(errno), nil
}

// Tell returns the position of the file
func (f *File) Tell() (ret int64, err error) {
	errno := C.int(C.lfs_file_tell(f.lfs.lfs, f.fileptr()))
	if errno < 0 {
		return -1, errval(errno)
	}
	return int64(errno), nil
}

// Rewind changes the position of the file to the beginning of the file
func (f *File) Rewind() (err error) {
	return errval(C.lfs_file_rewind(f.lfs.lfs, f.fileptr()))
}

// Size returns the size of the file
func (f *File) Size() (int64, error) {
	errno := C.int(C.lfs_file_size(f.lfs.lfs, f.fileptr()))
	if errno < 0 {
		return -1, errval(errno)
	}
	return int64(errno), nil
}

// Sync synchronizes to storage so that any pending writes are written out.
func (f *File) Sync() error {
	return errval(C.lfs_file_sync(f.lfs.lfs, f.fileptr()))
}

// Truncate the size of the file to the specified size
func (f *File) Truncate(size uint32) error {
	return errval(C.lfs_file_truncate(f.lfs.lfs, f.fileptr(), C.lfs_off_t(size)))
}

func (f *File) Write(buf []byte) (n int, err error) {
	bufptr := unsafe.Pointer(&buf[0])
	buflen := C.lfs_size_t(len(buf))
	errno := C.lfs_file_write(f.lfs.lfs, f.fileptr(), bufptr, buflen)
	if errno > 0 {
		return int(errno), nil
	} else {
		return 0, errval(C.int(errno))
	}
}

func (f *File) IsDir() bool {
	return f.typ == FileTypeDir
}

func (f *File) Readdir(n int) (infos []os.FileInfo, err error) {
	if n > 0 {
		return nil, errors.New("n > 0 is not supported yet")
	}
	if !f.IsDir() {
		return nil, ErrNotDir
	}
	for {
		var info C.struct_lfs_info
		i := C.lfs_dir_read(f.lfs.lfs, f.dirptr(), &info)
		if i == 0 {
			return
		}
		if i < 0 {
			err = errval(C.int(i))
			return
		}
		name := gostring(&info.name[0])
		if name == "." || name == ".." {
			continue // littlefs returns . and .., but Readdir() in Go does not
		}
		infos = append(infos, &Info{
			fileType: FileType(info._type),
			size:     uint32(info.size),
			name:     name,
		})
	}
}

// would be nice to use C.CString instead, but TinyGo doesn't seem to support
func cstring(s string) *C.char {
	ptr := C.malloc(C.size_t(len(s) + 1))
	buf := (*[1 << 28]byte)(ptr)[: len(s)+1 : len(s)+1]
	copy(buf, s)
	buf[len(s)] = 0
	return (*C.char)(ptr)
}

// would be nice to use C.GoString instead, but TinyGo doesn't seem to support
func gostring(s *C.char) string {
	slen := int(C.strlen(s))
	sbuf := make([]byte, slen)
	copy(sbuf, (*[1 << 28]byte)(unsafe.Pointer(s))[:slen:slen])
	return string(sbuf)
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
