// Places AppImage magic number and runtime sizes into a runtime file.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	aiMagic = []byte{0x41, 0x49, 0x02}
)

func main() {
	flag.Usage = func() {
		fmt.Println("go run stamp.go [runtime file] ...")
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		panic("Must provide a runtime file")
	}
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil {
			panic(err)
		}
		for _, r := range matches {
			f, err := os.OpenFile(r, os.O_WRONLY, 0755)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			_, err = f.WriteAt([]byte(aiMagic), 8)
			if err != nil {
				panic(err)
			}
			stat, err := os.Stat(r)
			if err != nil {
				panic(err)
			}
			siz := uint32(stat.Size())
			sizOut := make([]byte, 4)
			binary.LittleEndian.PutUint32(sizOut, siz)
			_, err = f.WriteAt(sizOut, 11)
			if err != nil {
				panic(err)
			}
		}
	}
}
