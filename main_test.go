package main

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/CalebQ42/squashfs"
)

var (
	appimageDownload = "https://github.com/srevinsaju/discord-appimage/releases/download/stable/Discord-0.0.22-x86_64.AppImage"
	appimage         = "testing/discord.appimage"
	appimageSquash   = "testing/discord.sfs"
	newAppImage      = "testing/test-discord.appimage"
)

// Creates a new file with the given AppImage squash archive with it as the runtime.
func TestStickon(t *testing.T) {
	setupTest(t)
	os.Remove(newAppImage)
	testFil, err := os.Create(newAppImage)
	if err != nil {
		t.Fatal(err)
	}
	defer testFil.Close()
	runtime, err := os.Open("static-appimage")
	if err != nil {
		t.Fatal(err)
	}
	n, err := io.Copy(testFil, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if n > runtimeSize {
		t.Fatal("runtime file is too big:", n)
	}
	fmt.Println("wrote", n)
	fmt.Println("padding", runtimeSize-n)
	_, err = testFil.Write(make([]byte, runtimeSize-n))
	if err != nil {
		t.Fatal(err)
	}
	runtime.Close()
	archive, err := os.Open(appimageSquash)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(testFil, archive)
	if err != nil {
		t.Fatal(err)
	}
	archive.Close()
	t.Fatal("success")
}

func TestMount(t *testing.T) {
	setupTest(t)
	exec.Command("umount", "testing/mountTest").Run()
	os.Mkdir("testing/mountTest", 0755)
	sfsFil, err := os.Open(appimageSquash)
	if err != nil {
		t.Fatal(err)
	}
	sfs, err := squashfs.NewReader(sfsFil)
	if err != nil {
		t.Fatal(err)
	}
	err = sfs.Mount("testing/mountTest")
	if err != nil {
		t.Fatal(err)
	}
	defer sfs.Unmount()
	sfs.MountWait()
}

func TestExtract(t *testing.T) {
	setupTest(t)
	exec.Command("umount", "testing/mountTest").Run()
	os.Mkdir("testing/mountTest", 0755)
	sfsFil, err := os.Open(appimageSquash)
	if err != nil {
		t.Fatal(err)
	}
	sfs, err := squashfs.NewReader(sfsFil)
	if err != nil {
		t.Fatal(err)
	}
	ops := squashfs.DefaultOptions()
	ops.Verbose = true
	err = sfs.ExtractWithOptions("testing/extractTest", ops)
	t.Fatal(err)
}

func setupTest(t *testing.T) {
	if _, err := os.Open(appimageSquash); os.IsNotExist(err) {
		var ai *os.File
		ai, err = os.Open(appimage)
		if os.IsNotExist(err) {
			ai, err = os.Create(appimage)
			if err != nil {
				t.Fatal(err)
			}
			defer ai.Close()
			var resp *http.Response
			resp, err = http.DefaultClient.Get(appimageDownload)
			if err != nil {
				os.Remove(appimage)
				t.Fatal(err)
			}
			defer resp.Body.Close()
			_, err = io.Copy(ai, resp.Body)
			if err != nil {
				os.Remove(appimage)
				t.Fatal(err)
			}
		}
		aiSfs, err := os.Create(appimageSquash)
		if err != nil {
			t.Fatal(err)
		}
		defer aiSfs.Close()
		off, err := CalculateElfSize(ai)
		if err != nil {
			t.Fatal(err)
		}
		stat, _ := ai.Stat()
		sec := io.NewSectionReader(ai, off, stat.Size()-off)
		_, err = io.Copy(aiSfs, sec)
		if err != nil {
			t.Fatal(err)
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
