package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CalebQ42/squashfs"
)

const (
	//
	runtimeSize = 4000000
)

func main() {
	fuse3 := true
	_, err := exec.LookPath("fusermount3")
	if err != nil {
		fuse3 = false
		_, err = exec.LookPath("fusermount")
		if err != nil {
			panic("Cannot mount AppImage, please check your FUSE setup.\nYou might still be able to extract the contents of this AppImage\nif you run it with the --appimage-extract option.\nSee https://github.com/AppImage/AppImageKit/wiki/FUSE\nfor more information")
		}
	}
	me := os.Args[0]
	me, err = filepath.Abs(me)
	if err != nil {
		panic(err)
	}
	meFil, err := os.Open(me)
	if err != nil {
		panic(err)
	}
	defer meFil.Close()
	sfs, err := squashfs.NewReaderAtOffset(meFil, runtimeSize)
	if err != nil {
		panic(err)
	}
	mntDir, err := os.MkdirTemp("", filepath.Base(me)+"-*")
	if err != nil {
		panic(err)
	}
	defer os.Remove(mntDir)
	if fuse3 {
		err = sfs.Mount(mntDir)
		if err != nil {
			panic(err)
		}
		defer sfs.Unmount()
	} else {
		err = sfs.MountFuse2(mntDir)
		if err != nil {
			panic(err)
		}
		defer sfs.UnmountFuse2()
	}
	cmd := exec.Command("sh", "-c", filepath.Join(mntDir, "AppRun"))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
