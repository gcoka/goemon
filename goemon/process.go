package goemon

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Process controls a command process.
type Process struct {
	cmdStr    string
	cmd       *exec.Cmd
	exit      chan error
	errStdout error
	errStderr error
	ExitCode  int
}

// NewProcess initializes Process.
func NewProcess(command string) *Process {
	p := &Process{}
	p.cmdStr = command
	p.exit = make(chan error)
	return p
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

	fmt.Println("PID:", p.cmd.Process.Pid)

	go func() {
		_, p.errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, p.errStderr = io.Copy(stderr, stderrIn)
	}()

	p.exit = make(chan error)

	go func() {
		err := p.cmd.Wait()
		if err != nil {
			log.Printf("cmd.Run() failed with %s", err)
		}
		s := p.cmd.ProcessState
		fmt.Println("string:", s.String())
		p.ExitCode = s.Sys().(syscall.WaitStatus).ExitStatus()
		fmt.Println("exitCode:", p.ExitCode)

		close(p.exit)
		p.cmd = nil
	}()

	return nil
}

// Interrupt sends interrupt signal to its children process.
func (p *Process) Interrupt() error {
	return syscall.Kill(-p.cmd.Process.Pid, syscall.SIGINT)
}

// Kill sends kill signal to its children process.
func (p *Process) Kill() error {
	return syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
}

// Stop kills command.
func (p *Process) Stop() error {
	if p.Exited() {
		fmt.Println("process is not running")
		return nil
	}

	if p.cmd != nil && p.cmd.Process != nil {
		if err := p.Interrupt(); err != nil {
			return err
		}

		select {
		case <-time.After(5 * time.Second):
			if err := p.Kill(); err != nil {
				log.Println("failed to kill: ", err)
			}
		case <-p.exit:
		}
	}

	return nil
}

// Wait waits till the command stops.
func (p *Process) Wait() error {
	if p.Exited() {
		return fmt.Errorf("Process is not running")
	}
	for {
		select {
		case <-p.exit:
			if p.errStdout != nil || p.errStderr != nil {
				return fmt.Errorf("failed to capture stdout or stderr")
			}
			return nil
		default:

		}
	}
}

// Restart stops current process and starts a new process.
func (p *Process) Restart() error {
	fmt.Println("restart begin")
	if !p.Exited() {
		p.Stop()
		err := p.Wait()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("stopped the old process")
	}
	time.Sleep(time.Second * 2)

	fmt.Println("start a new process")
	return p.Start()
}

// Exited returns if the command exited.
func (p *Process) Exited() bool {
	return p.cmd == nil || p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited()
}
