package flash_example

import (
	"fmt"
	"io"
	"machine"
	"os"
	"strconv"
	"strings"
	"time"

	lfs "github.com/bgould/go-littlefs"
	"tinygo.org/x/drivers/flash"
)

const consoleBufLen = 64
const storageBufLen = 512

var (
	debug = false

	input [consoleBufLen]byte
	store [storageBufLen]byte

	console  = machine.UART0
	readyLED = machine.LED

	flashdev *flash.Device
	blockdev lfs.BlockDevice
	fs       *lfs.LFS

	currdir = "/"

	commands = map[string]cmdfunc{
		"":        noop,
		"dbg":     dbg,
		"lsblk":   lsblk,
		"mount":   mount,
		"umount":  umount,
		"format":  format,
		"xxd":     xxd,
		"ls":      ls,
		"samples": samples,
		"mkdir":   mkdir,
		"cat":     cat,
		"create":  create,
		/*
			"cd":    cd,
		*/
	}
)

type cmdfunc func(argv []string)

const (
	StateInput = iota
	StateEscape
	StateEscBrc
	StateCSI
)

func RunFor(dev *flash.Device) {
	time.Sleep(3 * time.Second)

	flashdev = dev

	readyLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	readyLED.High()

	config := &flash.DeviceConfig{Identifier: flash.DefaultDeviceIdentifier}
	if err := dev.Configure(config); err != nil {
		for {
			time.Sleep(5 * time.Second)
			println("Config was not valid: "+err.Error(), "\r")
		}
	}

	readyLED.Low()
	println("SPI Configured. Reading flash info")

	lfsConfig := FlashLFSConfig()
	lfsConfig.BlockCount = flashdev.Attrs().TotalSize / lfsConfig.BlockSize
	println("Block count:", lfsConfig.BlockCount)

	blockdev = NewFlashBlockDevice(lfsConfig, flashdev)
	fs = lfs.New(lfsConfig, blockdev)

	prompt()

	var state = StateInput

	for i := 0; ; {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()
			if debug {
				fmt.Printf("\rdata: %x\r\n\r", data)
				prompt()
				console.Write(input[:i])
			}
			switch state {
			case StateInput:
				switch data {
				case 0x8:
					fallthrough
				case 0x7f: // this is probably wrong... works on my machine tho :)
					// backspace
					if i > 0 {
						i -= 1
						console.Write([]byte{0x8, 0x20, 0x8})
					}
				case 13:
					// return key
					console.Write([]byte("\r\n"))
					runCommand(string(input[:i]))
					prompt()

					i = 0
					continue
				case 27:
					// escape
					state = StateEscape
				default:
					// anything else, just echo the character if it is printable
					if strconv.IsPrint(rune(data)) {
						if i < (consoleBufLen - 1) {
							console.WriteByte(data)
							input[i] = data
							i++
						}
					}
				}
			case StateEscape:
				switch data {
				case 0x5b:
					state = StateEscBrc
				default:
					state = StateInput
				}
			default:
				// TODO: handle escape sequences
				state = StateInput
			}
			//time.Sleep(10 * time.Millisecond)
		}
	}
}

func runCommand(line string) {
	argv := strings.SplitN(strings.TrimSpace(line), " ", -1)
	cmd := argv[0]
	cmdfn, ok := commands[cmd]
	if !ok {
		println("unknown command: " + line)
		return
	}
	cmdfn(argv)
}

func noop(argv []string) {}

func dbg(argv []string) {
	if debug {
		debug = false
		println("Console debugging off")
	} else {
		debug = true
		println("Console debugging on")
	}
}

func lsblk(argv []string) {
	attrs := flashdev.Attrs()
	status1, _ := flashdev.ReadStatus()
	status2, _ := flashdev.ReadStatus2()
	serialNumber1, _ := flashdev.ReadSerialNumber()
	fmt.Printf(
		"\n-------------------------------------\r\n"+
			" Device Information:  \r\n"+
			"-------------------------------------\r\n"+
			" JEDEC ID: %v\r\n"+
			" Serial:   %v\r\n"+
			" Status 1: %02x\r\n"+
			" Status 2: %02x\r\n"+
			" \r\n"+
			" Max clock speed (MHz): %d\r\n"+
			" Has Sector Protection: %t\r\n"+
			" Supports Fast Reads:   %t\r\n"+
			" Supports QSPI Reads:   %t\r\n"+
			" Supports QSPI Write:   %t\r\n"+
			" Write Status Split:    %t\r\n"+
			" Single Status Byte:    %t\r\n"+
			"-------------------------------------\r\n\r\n",
		attrs.JedecID,
		serialNumber1,
		status1,
		status2,
		attrs.MaxClockSpeedMHz,
		attrs.HasSectorProtection,
		attrs.SupportsFastRead,
		attrs.SupportsQSPI,
		attrs.SupportsQSPIWrites,
		attrs.WriteStatusSplit,
		attrs.SingleStatusByte,
	)
}

func mount(argv []string) {
	if err := fs.Mount(); err != nil {
		println("Could not mount LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully mounted LittleFS filesystem.\r\n")
	}
}

func format(argv []string) {
	if err := fs.Format(); err != nil {
		println("Could not format LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully formatted LittleFS filesystem.\r\n")
	}
}

func umount(argv []string) {
	if err := fs.Unmount(); err != nil {
		println("Could not unmount LittleFS filesystem: " + err.Error() + "\r\n")
	} else {
		println("Successfully unmounted LittleFS filesystem.\r\n")
	}
}

/*
	var err error
	if fatfs == nil {
		fatfs, err = fat.New(fatdisk)
		if err != nil {
			fatfs = nil
			println("could not load FAT filesystem: " + err.Error() + "\r\n")
		}
		fmt.Printf("loaded fs\r\n")
	}
	if rootdir == nil {
		rootdir, err = fatfs.RootDir()
		if err != nil {
			rootdir = nil
			println("could not load rootdir: " + err.Error() + "\r\n")
		}
		fmt.Printf("loaded rootdir\r\n")
	}
	if currdir == nil {
		currdir = rootdir
	}
*/

func ls(argv []string) {
	path := "/"
	if len(argv) > 1 {
		path = strings.TrimSpace(argv[1])
	}
	dir, err := fs.Open(path)
	if err != nil {
		fmt.Printf("Could not open directory %s: %v\n", path, err)
		return
	}
	defer dir.Close()
	infos, err := dir.Readdir(0)
	_ = infos
	if err != nil {
		fmt.Printf("Could not read directory %s: %v\n", path, err)
		return
	}
	for _, info := range infos {
		s := "-rwxrwxrwx"
		if info.IsDir() {
			s = "drwxrwxrwx"
		}
		fmt.Printf("%s %5d %s\n", s, info.Size(), info.Name())
	}
}

func mkdir(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying mkdir to " + tgt)
	}
	if tgt == "" {
		println("Usage: mkdir <target dir>")
		return
	}
	err := fs.Mkdir(tgt)
	if err != nil {
		println("Could not mkdir " + tgt + ": " + err.Error())
	}
}

func samples(argv []string) {
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("file%d.txt", i)
		if bytes, err := createSampleFile(name); err != nil {
			fmt.Printf("%s\r\n", err)
			return
		} else {
			fmt.Printf("wrote %d bytes to %s\r\n", bytes, name)
		}
	}
}

func create(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying create to " + tgt)
	}
	if bytes, err := createSampleFile(tgt); err != nil {
		fmt.Printf("%s\r\n", err)
		return
	} else {
		fmt.Printf("wrote %d bytes to %s\r\n", bytes, tgt)
	}
}

func createSampleFile(name string) (int, error) {
	buf := make([]byte, 90)
	for j := uint8(0); j < uint8(len(buf)); j++ {
		buf[j] = 0x20 + j
	}
	f, err := fs.OpenFile(name, lfs.O_CREAT|lfs.O_WRONLY|lfs.O_TRUNC)
	if err != nil {
		return 0, fmt.Errorf("error opening %s: %s", name, err.Error())
	}
	defer f.Close()
	bytes, err := f.Write(buf)
	if err != nil {
		return 0, fmt.Errorf("error writing %s: %s", name, err.Error())
	}
	return bytes, nil
}

/*
func cd(argv []string) {

	if fatfs == nil || rootdir == nil {
		mnt(nil)
	}
	if len(argv) == 1 {
		currdir = rootdir
		return
	}
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying to cd to " + tgt)
	}
	if tgt == "" {
		println("Usage: cd <target dir>")
		return
	}
	if debug {
		println("Getting entry")
	}
	entry := currdir.Entry(tgt)
	if entry == nil {
		println("File not found: " + tgt)
		return
	}
	if !entry.IsDir() {
		println("Not a directory: " + tgt)
		return
	}
	if debug {
		println("Getting dir")
	}
	cd, err := entry.Dir()
	if err != nil {
		println("Could not cd to " + tgt + ": " + err.Error())
	}
	currdir = cd
}
*/

func cat(argv []string) {
	tgt := ""
	if len(argv) == 2 {
		tgt = strings.TrimSpace(argv[1])
	}
	if debug {
		println("Trying to cat to " + tgt)
	}
	if tgt == "" {
		println("Usage: cat <target dir>")
		return
	}
	if debug {
		println("Getting entry")
	}
	f, err := fs.Open(tgt)
	if err != nil {
		println("Could not open: " + err.Error())
		return
	}
	defer f.Close()
	if f.IsDir() {
		println("Not a file: " + tgt)
		return
	}
	off := 0x0
	buf := make([]byte, 64)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			println("Error reading " + tgt + ": " + err.Error())
		}
		xxdfprint(os.Stdout, uint32(off), buf[:n])
		off += n
	}
}

func xxd(argv []string) {
	var err error
	var addr uint64 = 0x0
	var size int = 64
	switch len(argv) {
	case 3:
		if size, err = strconv.Atoi(argv[2]); err != nil {
			println("Invalid size argument: " + err.Error() + "\r\n")
			return
		}
		if size > storageBufLen || size < 1 {
			fmt.Printf("Size of hexdump must be greater than 0 and less than %d\r\n", storageBufLen)
			return
		}
		fallthrough
	case 2:
		/*
			if argv[1][:2] != "0x" {
				println("Invalid hex address (should start with 0x)")
				return
			}
		*/
		if addr, err = strconv.ParseUint(argv[1], 16, 32); err != nil {
			println("Invalid address: " + err.Error() + "\r\n")
			return
		}
		fallthrough
	case 1:
		// no args supplied, so nothing to do here, just use the defaults
	default:
		println("usage: xxd <hex address, ex: 0xA0> <size of hexdump in bytes>\r\n")
		return
	}
	buf := store[0:size]
	bsz := uint64(flash.BlockSize)
	blockdev.ReadBlock(uint32(addr/bsz), uint32(addr%bsz), buf)
	//fatdisk.ReadAt(buf, int64(addr))
	xxdfprint(os.Stdout, uint32(addr), buf)
}

func xxdfprint(w io.Writer, offset uint32, b []byte) {
	var l int
	var buf16 = make([]byte, 16)
	var padding = ""
	for i, c := 0, len(b); i < c; i += 16 {
		l = i + 16
		if l >= c {
			padding = strings.Repeat(" ", (l-c)*3)
			l = c
		}
		fmt.Fprintf(w, "%08x: % x    "+padding, offset+uint32(i), b[i:l])
		for j, n := 0, l-i; j < 16; j++ {
			if j >= n {
				buf16[j] = ' '
			} else if !strconv.IsPrint(rune(b[i+j])) {
				buf16[j] = '.'
			} else {
				buf16[j] = b[i+j]
			}
		}
		console.Write(buf16)
		println()
		//	"%s\r\n", b[i:l], "")
	}
}

func prompt() {
	print("==> ")
}
