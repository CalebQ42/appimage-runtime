package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CalebQ42/squashfs"
)

func usage() {
	fmt.Println(filepath.Base(os.Args[0]), "[--appimage-extract]", "[-m <Mount location>]")
	fmt.Println("")
	flag.PrintDefaults()
}

func main() {
	otherE := flag.Bool("appimage-extract", false, "Extract the AppImage archive to $FILENAME.extract")
	e := flag.Bool("e", false, "Extract the AppImage archive to $FILENAME.extract Cannot combine with -m.")
	mnt := flag.String("m", "", "Mount the AppImage to the given location")
	flag.Usage = usage
	flag.Parse()
	extract := *e || *otherE
	fuse3 := true
	if !extract {
		_, err := exec.LookPath("fusermount3")
		if err != nil {
			fuse3 = false
			_, err = exec.LookPath("fusermount")
			if err != nil {
				panic("Cannot mount AppImage, please check your FUSE setup.\nYou might still be able to extract the contents of this AppImage\nif you run it with the --appimage-extract option.\nSee https://github.com/AppImage/AppImageKit/wiki/FUSE\nfor more information")
			}
		}
	}
	me := os.Args[0]
	me, err := filepath.Abs(me)
	if err != nil {
		panic(err)
	}
	meFil, err := os.Open(me)
	if err != nil {
		panic(err)
	}
	defer meFil.Close()
	var runtimeSize int32
	err = binary.Read(io.NewSectionReader(meFil, 11, 4), binary.LittleEndian, &runtimeSize)
	if err != nil {
		panic(err)
	}
	sfs, err := squashfs.NewReaderAtOffset(meFil, int64(runtimeSize))
	if err != nil {
		panic(err)
	}
	if extract {
		os.Remove(me + ".extract")
		err = sfs.ExtractTo(me + ".extract")
		if err != nil {
			panic(err)
		}
		return
	}
	if *mnt != "" {
		fmt.Println("Please run \"umount " + *mnt + "\" when you are done")
		if fuse3 {
			err = sfs.Mount(*mnt)
			if err != nil {
				panic(err)
			}
			fmt.Println("Please run \"umount " + *mnt + "\" when you are done")
			sfs.MountWait()
		} else {
			err = sfs.MountFuse2(*mnt)
			if err != nil {
				panic(err)
			}
			sfs.MountWaitFuse2()
		}
		return
	}
	mntDir, err := os.MkdirTemp("", "appimage*")
	if err != nil {
		panic(err)
	}
	defer os.Remove(mntDir)
	if fuse3 {
		err = sfs.Mount(mntDir)
		if err != nil {
			panic(err)
		}
	} else {
		err = sfs.MountFuse2(mntDir)
		if err != nil {
			panic(err)
		}
	}
	// TODO: Change to sfs.Unmount.
	// There is some sort of race condition in the fuse library that causes it to not work and hang indefinitely.
	defer exec.Command("umount", mntDir).Run()
	cmd := exec.Command("sh", "-c", filepath.Join(mntDir, "AppRun"))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
