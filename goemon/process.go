package goemon

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Process controls a command process.
type Process struct {
	cmdStr     string
	cmd        *exec.Cmd
	exit       chan error
	errStdout  error
	errStderr  error
	exitCode   int
	pid        int
	verbose    bool
	restarting chan int
	started    time.Time
}

// NewProcess initializes Process.
func NewProcess(command string) *Process {
	p := &Process{}
	p.cmdStr = command
	p.exit = make(chan error)
	p.restarting = make(chan int, 2)
	return p
}

// SetVerbose sets verbose option.
func (p *Process) SetVerbose(v bool) {
	p.verbose = v
}

// ExitCode returns the process id of the running command.
func (p *Process) ExitCode() int {
	return p.exitCode
}

// PID is the process id of the running command.
func (p *Process) PID() int {
	return p.pid
}

// Started returns started time.
func (p *Process) Started() time.Time {
	return p.started
}

// Start starts a command and wait to end.
func (p *Process) Start() error {

	if !p.Exited() {
		return fmt.Errorf("process is running")
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("sh", "-c", p.cmdStr)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	p.cmd = cmd

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd.Start() failed with '%s'", err)
	}

	p.pid = p.cmd.Process.Pid

	if p.verbose {
		fmt.Println(p.String())
	}

	go func() {
		_, p.errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, p.errStderr = io.Copy(stderr, stderrIn)
	}()

	p.exit = make(chan error)
	p.started = time.Now()

	go func() {
		err := p.cmd.Wait()
		if err != nil {
			fmt.Printf("cmd.Run() failed with %s\n", err)
		}
		s := p.cmd.ProcessState
		p.exitCode = s.Sys().(syscall.WaitStatus).ExitStatus()

		if p.verbose {
			fmt.Println("process exited with", p.exitCode)
		}

		close(p.exit)
		p.cmd = nil
	}()

	return nil
}

// Interrupt sends interrupt signal to its children process.
func (p *Process) Interrupt() error {
	if p.verbose {
		fmt.Println("Send interrupt")
	}
	return syscall.Kill(-p.cmd.Process.Pid, syscall.SIGINT)
}

// Kill sends kill signal to its children process.
func (p *Process) Kill() error {
	return syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
}

// Stop kills command.
func (p *Process) Stop() error {
	if p.verbose {
		fmt.Println("Stop process invoked")
	}

	if p.Exited() {
		return nil
	}

	if p.cmd != nil && p.cmd.Process != nil {
		if err := p.Interrupt(); err != nil {
			return err
		}

		select {
		case <-time.After(5 * time.Second):
			if err := p.Kill(); err != nil {
				return fmt.Errorf("failed to kill: %v", err)
			}
		case <-p.exit:
		}
	}

	return nil
}

// Wait waits till the command stops.
func (p *Process) Wait() error {
	if p.Exited() {
		return fmt.Errorf("Wait called but Process is not running")
	}

	<-p.exit

	if p.errStdout != nil || p.errStderr != nil {
		if p.verbose {
			fmt.Println("failed to capture stdout or stderr", p.errStdout, p.errStderr)
		}
	}
	return nil
}

// Restart stops current process and starts a new process.
func (p *Process) Restart() error {
	fmt.Println("[debug] restart invoked")
	if len(p.restarting) > 0 {
		fmt.Println("[debug] restarting")
		return fmt.Errorf("restarting")
	}
	p.restarting <- 1
	if !p.Exited() {
		p.Stop()
		err := p.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}
	err := p.Start()
	<-p.restarting

	fmt.Println("successfully restarted")
	return err
}

// Exited returns if the command exited.
func (p *Process) Exited() bool {
	return p.cmd == nil || p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited()
}

func (p *Process) String() string {
	return fmt.Sprintf("[PID: %v] %v", p.PID(), p.cmdStr)
}
