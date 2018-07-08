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

// StartCloser provides start and close.
type StartCloser interface {
	Start() error
	Close()
}

// Option has option parameters.
type Option struct {
	Delay        int
	Ext          []string
	Watches      []string
	Ignores      []string
	PrintWatches bool
	Verbose      bool
}

// Default sets default option values.
func (o *Option) Default() {
	if o.Delay == 0 {
		o.Delay = 500
	}
	if o.Ext == nil {
		o.Ext = []string{}
	}
	if o.Watches == nil {
		o.Watches = []string{"."}
	}
	if o.Ignores == nil {
		o.Ignores = []string{}
	}
	o.Ignores = append(o.Ignores, ".git", ".git/**")
}

// NormalizeExt normalize comma-separated or space-separated extentions.
// like ["go,md", "yml json"] into single ext valued array ["go", "md", "yml", "json"].
func NormalizeExt(ext []string) []string {
	n := make(map[string]struct{})
	seps := []string{",", " "}
	for _, v := range ext {
		for _, s := range seps {
			for _, e := range strings.Split(v, s) {
				n[e] = struct{}{}
			}
		}
	}

	keys := make([]string, 0, len(n))
	for k := range n {
		if strings.ContainsAny(k, strings.Join(seps, "")) {
			continue
		}
		keys = append(keys, k)
	}
	return keys
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
	watcher    *watcher.Watcher
	watchStart time.Time
	processes  []*Process
	option     *Option
	watches    []glob.Glob
	ignores    []glob.Glob
}

// New initializes Goemon watcher.
func New(cmds []string, opt *Option) StartCloser {
	if opt == nil {
		opt = &Option{}
	}
	opt.Default()

	opt.Ext = NormalizeExt(opt.Ext)

	procs := make([]*Process, 0, len(cmds))
	for _, v := range cmds {
		p := NewProcess(v)
		p.SetVerbose(opt.Verbose)
		procs = append(procs, p)
	}

	return &Goemon{
		processes: procs,
		option:    opt,
		watches:   CompileGlobs(opt.Watches),
		ignores:   CompileGlobs(opt.Ignores),
	}
}

func newWatcher() *watcher.Watcher {
	w := watcher.New()

	w.SetMaxEvents(10)
	w.IgnoreHiddenFiles(false)

	w.FilterOps(
		watcher.Remove,
		watcher.Write,
		watcher.Rename,
		watcher.Move,
		watcher.Chmod,
		watcher.Create,
	)
	return w
}

// ListTarget lists files accouding to watches and ignores globbing pattern.
func ListTarget(watches, ignores []glob.Glob) map[string]os.FileInfo {

	targets := make(map[string]os.FileInfo)

	wWalker := NewGlobWalker(watches)

	wWalker.Walk(".", func(target string, fi os.FileInfo, e error) error {
		targets[target] = fi
		return nil
	})

	iWalker := NewGlobWalker(ignores)

	iWalker.Walk(".", func(ignore string, fi os.FileInfo, e error) error {
		if fi.IsDir() {
			p := ignore + string(os.PathSeparator) + "**"
			w := NewGlobWalker(CompileGlobs([]string{p}))
			w.Walk(ignore, func(f string, cfi os.FileInfo, ce error) error {
				delete(targets, f)
				return nil
			})

		} else {
			p := "**" + trimPathSeparator(ignore)
			w := NewGlobWalker(CompileGlobs([]string{p}))
			w.Walk(ignore, func(f string, cfi os.FileInfo, ce error) error {
				delete(targets, f)
				return nil
			})
		}

		delete(targets, ignore)
		return nil
	})

	return targets
}

func trimPathSeparator(s string) string {
	sep := string(os.PathSeparator)
	return strings.Trim(s, sep)
}

// Start starts watching.
func (g *Goemon) Start() error {

	g.watcher = newWatcher()

	for k := range ListTarget(g.watches, g.ignores) {
		g.watcher.Add(k)
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
				if event.ModTime().Before(
					g.watchStart.Add(time.Duration(g.option.Delay) * time.Millisecond),
				) {
					continue
				}
				if g.option.Verbose {
					fmt.Println(event.ModTime(), event) // Print the event's info.
				}
				ext := filepath.Ext(event.Path)
				if ext == "" || g.option.IsTargetExt(ext) {
					for _, p := range g.processes {
						if event.ModTime().Before(
							p.Started().Add(time.Duration(g.option.Delay) * time.Millisecond),
						) {
							continue
						}
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
	g.watchStart = time.Now()
	return g.watcher.Start(time.Millisecond * 200)
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
	files := g.listWatchedFiles()
	fmt.Println(files)
}

func (g *Goemon) listWatchedFiles() []string {
	files := make([]string, 0, len(g.watcher.WatchedFiles()))
	cwd, _ := os.Getwd()
	for k := range g.watcher.WatchedFiles() {
		f, _ := filepath.Rel(cwd, k)
		files = append(files, f)
	}
	return files
}
