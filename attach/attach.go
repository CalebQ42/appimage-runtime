package main

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
)

func main() {
	runtimeFil := flag.String("r", "", "The runtime file to attach")
	strip := flag.Bool("s", false, "Strip the AppImage and save it before attaching the new runtime")
	flag.Parse()
	if *runtimeFil == "" {
		panic("Must specify -r [runtime file]")
	}
	dirs, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}
	for _, d := range dirs {
		if strings.HasSuffix(strings.ToLower(d.Name()), ".appimage") {
			name := d.Name()
			name = name[:len(name)-9]
			ai, err := os.Open(d.Name())
			if err != nil {
				panic(err)
			}
			aiStat, _ := ai.Stat()
			offset, err := CalculateElfSize(ai)
			if err != nil {
				panic(err)
			}
			if *strip {
				var aiSfs, sfs *os.File
				aiSfs, err = os.Open(d.Name())
				if err != nil {
					panic(err)
				}
				sec := io.NewSectionReader(aiSfs, offset, aiStat.Size()-offset)
				sfs, err = os.Create(name + ".sfs")
				if err != nil {
					panic(err)
				}
				_, err = io.Copy(sfs, sec)
				if err != nil {
					panic(err)
				}
			}
			runtime, err := os.Open(*runtimeFil)
			if err != nil {
				panic(err)
			}
			newFil, err := os.Create(name + "-new.AppImage")
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(newFil, runtime)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(newFil, io.NewSectionReader(ai, offset, aiStat.Size()-offset))
			if err != nil {
				panic(err)
			}
			newFil.Chmod(0755)
			newFil.Close()
			runtime.Close()
			ai.Close()
		}
	}
}

// CalculateElfSize returns the size of an ELF binary as an int64 based on the information in the ELF header.
// Taken from https://github.com/probonopd/go-appimage
func CalculateElfSize(f *os.File) (int64, error) {

	e, err := elf.NewFile(f)
	if err != nil {
		return 0, err
	}

	// Read identifier
	var ident [16]uint8
	_, err = f.ReadAt(ident[0:], 0)
	if err != nil {
		return 0, err
	}

	// Decode identifier
	if ident[0] != '\x7f' ||
		ident[1] != 'E' ||
		ident[2] != 'L' ||
		ident[3] != 'F' {
		return 0, errors.New("bad magic number")
	}
	// Process by architecture
	sr := io.NewSectionReader(f, 0, 1<<63-1)
	var shoff, shentsize, shnum int64
	switch e.Class.String() {
	case "ELFCLASS64":
		hdr := new(elf.Header64)
		_, err = sr.Seek(0, 0)
		if err != nil {
			return 0, err
		}
		err = binary.Read(sr, e.ByteOrder, hdr)
		if err != nil {
			return 0, err
		}

		shoff = int64(hdr.Shoff)
		shnum = int64(hdr.Shnum)
		shentsize = int64(hdr.Shentsize)
	case "ELFCLASS32":
		hdr := new(elf.Header32)
		_, err = sr.Seek(0, 0)
		if err != nil {
			return 0, err
		}
		err = binary.Read(sr, e.ByteOrder, hdr)
		if err != nil {
			return 0, err
		}

		shoff = int64(hdr.Shoff)
		shnum = int64(hdr.Shnum)
		shentsize = int64(hdr.Shentsize)
	default:
		return 0, errors.New("unsupported elf architecuture")
	}

	// Calculate ELF size
	elfsize := shoff + (shentsize * shnum)
	// log.Println("elfsize:", elfsize, file)
	return elfsize, nil
}
