package goemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/radovskyb/watcher"
)

// Option has option parameters.
type Option struct {
	WatchInterval int
	Ext           []string
	Watches       []string
	Ignores       []string
	PrintWatches  bool
}

// Default sets default option values.
func (o *Option) Default() {
	if o.WatchInterval == 0 {
		o.WatchInterval = 500
	}
	if o.Ext == nil {
		o.Ext = []string{}
	}
	if o.Watches == nil {
		o.Watches = []string{"."}
	}
	if o.Ignores == nil {
		o.Ignores = []string{"."}
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

// Goemon is file monitor.
type Goemon struct {
	watcher   *watcher.Watcher
	processes []*Process
	option    *Option
	watches   []glob.Glob
	ignores   []glob.Glob
}

// New initializes Goemon watcher.
func New(cmds []string, opt *Option) *Goemon {
	if opt == nil {
		opt = &Option{}
	}
	opt.Default()

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
		watches:   CompileGlobs(opt.Watches),
		ignores:   CompileGlobs(opt.Ignores),
	}
}

// Start starts watching.
func (g *Goemon) Start() error {

	wWalker := NewGlobWalker(g.watches)

	for _, target := range wWalker.Walk(".") {
		err := g.watcher.AddRecursive(target)
		if err != nil {
			return err
		}
	}

	iWalker := NewGlobWalker(g.ignores)

	for _, ignore := range iWalker.Walk(".") {
		err := g.watcher.Ignore(ignore)
		if err != nil {
			return err
		}
	}

	for _, p := range g.processes {
		err := p.Start()
		if err != nil {
			fmt.Println(err)
		}
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

	if g.option.PrintWatches {
		g.PrintWatchedFiles()
	}
	return g.watcher.Start(time.Millisecond * time.Duration(g.option.WatchInterval))
}

// Close stops watching.
func (g *Goemon) Close() {
	g.watcher.Close()
	for _, p := range g.processes {
		p.Stop()
	}
}

// PrintWatchedFiles prints
func (g *Goemon) PrintWatchedFiles() {
	files := g.listWatchFiles()
	fmt.Println(files)
}

func (g *Goemon) listWatchFiles() []string {
	files := make([]string, 0, len(g.watcher.WatchedFiles()))
	cwd, _ := os.Getwd()
	for k := range g.watcher.WatchedFiles() {
		f, _ := filepath.Rel(cwd, k)
		files = append(files, f)
	}
	return files
}
