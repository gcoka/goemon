package goemon

import (
	"fmt"
	"log"
	"time"

	"github.com/radovskyb/watcher"
)

type GoemonOption struct {
}

type Goemon struct {
	watcher *watcher.Watcher
}

func New() *Goemon {
	w := watcher.New()

	w.SetMaxEvents(3)
	w.IgnoreHiddenFiles(false)

	w.FilterOps(watcher.Remove,
		watcher.Write,
		watcher.Rename,
		watcher.Move,
		watcher.Chmod,
		watcher.Create,
	)

	return &Goemon{
		watcher: w,
	}
}

func (g *Goemon) Start() error {

	err := g.watcher.AddRecursive(".")
	if err != nil {
		return err
	}

	watch := func(w *watcher.Watcher) {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event) // Print the event's info.
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}

	go watch(g.watcher)

	err = g.watcher.Start(time.Millisecond * 500)
	return err
}

func (g *Goemon) Close() {
	g.watcher.Close()
}
