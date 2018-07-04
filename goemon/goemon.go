package goemon

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)

// Option has option parameters.
type Option struct {
	WatchInterval int
	Ext           []string
}

// Default sets default option values.
func (o *Option) Default() {
	if o.WatchInterval == 0 {
		o.WatchInterval = 500
	}
	if o.Ext == nil {
		o.Ext = []string{}
	}
}

// IsTargetExt detects given ext is in option.
func (o *Option) IsTargetExt(ext string) bool {
	for _, v := range o.Ext {
		if strings.TrimPrefix(v, ".") == strings.TrimPrefix(ext, ".") {
			return true
		}
	}
	return false
}

// Goemon is
type Goemon struct {
	watcher   *watcher.Watcher
	processes []*Process
	option    *Option
}

// New initializes Goemon watcher.
func New(cmds []string, opt *Option) *Goemon {
	if opt == nil {
		opt = &Option{}
	}
	opt.Default()

	if opt.Ext != nil {
		fmt.Println(opt.Ext)
	}

	procs := make([]*Process, 0, len(cmds))
	for _, v := range cmds {
		procs = append(procs, NewProcess(v))
	}

	w := watcher.New()

	w.SetMaxEvents(10000)
	w.IgnoreHiddenFiles(false)

	w.FilterOps(
		watcher.Remove,
		watcher.Write,
		watcher.Rename,
		watcher.Move,
		watcher.Chmod,
		watcher.Create,
	)

	return &Goemon{
		processes: procs,
		watcher:   w,
		option:    opt,
	}
}

// Start starts watching.
func (g *Goemon) Start() error {

	for _, p := range g.processes {
		err := p.Start()
		if err != nil {
			fmt.Println(err)
		}
	}

	err := g.watcher.AddRecursive(".")
	if err != nil {
		return err
	}

	watch := func(g *Goemon) {
		for {
			select {
			case event := <-g.watcher.Event:
				fmt.Println(event) // Print the event's info.
				ext := filepath.Ext(event.Path)
				if g.option.IsTargetExt(ext) {
					for _, p := range g.processes {
						err := p.Restart()
						if err != nil {
							fmt.Println(err)
						}
					}
				}

			case err := <-g.watcher.Error:
				log.Fatalln(err)
			case <-g.watcher.Closed:
				fmt.Println("watcher closed.")
				return
			default:
			}
		}
	}

	go watch(g)
	fmt.Println(g.watcher.WatchedFiles())
	fmt.Println("watch interval", g.option.WatchInterval)
	err = g.watcher.Start(time.Millisecond * time.Duration(g.option.WatchInterval))
	return err
}

// Close stops watching.
func (g *Goemon) Close() {
	g.watcher.Close()
	for _, p := range g.processes {
		p.Stop()
	}
}
