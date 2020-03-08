package main

import (
	"machine"

	fatfs_example "github.com/bgould/sandbox/flash-fatfs"

	"tinygo.org/x/drivers/flash"
)

func main() {
	fatfs_example.RunFor(
		flash.NewSPI(
			&machine.SPI1,
			machine.SPI1_MOSI_PIN,
			machine.SPI1_MISO_PIN,
			machine.SPI1_SCK_PIN,
			machine.SPI1_CS_PIN,
		),
	)
}
