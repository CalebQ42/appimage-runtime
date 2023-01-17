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
	sfs, err := squashfs.NewReaderAtOffset(meFil, runtimeSize)
	if err != nil {
		panic(err)
	}
	mntDir, err := os.MkdirTemp("", filepath.Base(me)+"-*")
	if err != nil {
		panic(err)
	}
	defer os.Remove(mntDir)
	err = sfs.Mount(mntDir)
	if err != nil {
		panic(err)
	}
	defer sfs.Unmount()
	cmd := exec.Command("sh", "-c", filepath.Join(mntDir, "AppRun"))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
