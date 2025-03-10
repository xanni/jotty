package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xanni/jotty/edits"
	"github.com/xanni/jotty/i18n"
	ps "github.com/xanni/jotty/permascroll"
)

//go:generate sh -c "printf %s $(git describe --always --tags) > version.txt"
//go:embed version.txt
var version string

const (
	defaultExport      = "jotty.txt"
	defaultPermascroll = "jotty.jot"
)

func cleanup() {
	ps.Flush()
	if err := ps.ClosePermascroll(); err != nil {
		log.Printf("%+v", err)
	}
}

func usage() {
	fmt.Println("https://github.com/xanni/jotty  ⓒ 2024–2025 Andrew Pam <xanni@xanadu.net>")
	fmt.Printf("\n"+i18n.Text["usage"]+"\n", filepath.Base(os.Args[0]), defaultPermascroll)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	vFlag := flag.Bool("version", false, i18n.Text["version"])
	flag.Parse()
	if *vFlag {
		println(filepath.Base(os.Args[0]) + " " + version)
		os.Exit(0)
	}

	exportPath, permascrollPath := defaultExport, defaultPermascroll
	if len(os.Args) > 1 {
		exportPath, permascrollPath = flag.Arg(0), flag.Arg(0)
		if i := strings.LastIndex(exportPath, ".jot"); i >= 0 {
			exportPath = exportPath[:i]
		}
		exportPath += ".txt"
	}

	if err := ps.OpenPermascroll(permascrollPath); err != nil {
		log.Fatalf("%+v", err)
	}

	defer cleanup()

	edits.Run(version, exportPath)
}
