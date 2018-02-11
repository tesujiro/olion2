package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/tesujiro/olion2"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "Error:\n%s", err)
			os.Exit(1)
		}
	}()
	os.Exit(_main())
}

type ignorable interface {
	Ignorable() bool
}

type causer interface {
	Cause() error
}

func _main() int {
	if envvar := os.Getenv("GOMAXPROCS"); envvar == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	olion.InitColor()

	termbox.SetOutputMode(termbox.Output256)
	//termbox.SetOutputMode(termbox.OutputGrayscale)

	olion := olion.New(ctx, cancel)

	cpuprofile := "mycpu.prof"
	f, err := os.Create(cpuprofile)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	if err := olion.Run(ctx); err != nil {
		for e := err; e != nil; {
			switch e.(type) {
			case ignorable:
				time.Sleep(3 * time.Second)
				return 0
			case causer:
				e = e.(causer).Cause()
			}
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	return 0
}
