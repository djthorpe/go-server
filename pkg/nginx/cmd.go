package nginx

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cmd struct {
	cmd         *exec.Cmd
	Out, Err    CallbackFn
	Start, Stop time.Time
}

// Callback output from the command. Newlines are embedded
// within the string
type CallbackFn func(*Cmd, []byte)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithCommand(cmd string, args ...string) (*Cmd, error) {
	this := new(Cmd)
	if !filepath.IsAbs(cmd) {
		if cmd_, err := exec.LookPath(cmd); err != nil {
			return nil, err
		} else {
			cmd = cmd_
		}
	}
	if stat, err := os.Stat(cmd); err != nil {
		return nil, err
	} else if !IsExecAny(stat.Mode()) {
		return nil, ErrBadParameter.Withf("Command is not executable: %q", cmd)
	} else {
		this.cmd = exec.Command(cmd, args...)
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Cmd) String() string {
	str := "<cmd"
	if t.cmd != nil {
		str += fmt.Sprintf(" exec=%q", t.cmd.Path)
		if len(t.cmd.Args) > 1 {
			str += fmt.Sprintf(" args=%q", t.cmd.Args[1:])
		}
		for _, v := range t.cmd.Env {
			str += fmt.Sprint(v)
		}
		if t.cmd.Process != nil {
			if pid := t.cmd.Process.Pid; pid > 0 {
				str += fmt.Sprintf(" pid=%d", pid)
			}
			if t.cmd.ProcessState.Exited() {
				str += fmt.Sprintf(" exit_code=%d", t.cmd.ProcessState.ExitCode())
			}
		}
	}
	if !t.Start.IsZero() && !t.Stop.IsZero() {
		str += fmt.Sprintf(" duration=%q", t.Stop.Sub(t.Start).Truncate(time.Second))
	} else if !t.Start.IsZero() {
		str += fmt.Sprintf(" since=%q", t.Start.Format(time.Kitchen))
	}

	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run the command in the foreground, and return any errors
func (c *Cmd) Run() error {
	var wg sync.WaitGroup

	// Check to see if command has alreay been run
	if c.cmd.Process != nil {
		return ErrOutOfOrder.With("Command has already been run")
	}

	// Pipes for reading stdout and stderr
	if c.Out != nil {
		stdout, err := c.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.read(stdout, c.Out)
		}()
	}
	if c.Err != nil {
		stderr, err := c.cmd.StderrPipe()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.read(stderr, c.Err)
		}()
	}

	// Start command, mark start and stop times
	c.Start = time.Now()
	defer func() {
		c.Stop = time.Now()
	}()
	if err := c.cmd.Start(); err != nil {
		return err
	}

	// Wait for stdout and stderr to be closed
	wg.Wait()

	// Wait for command to exit, return any errors
	return c.cmd.Wait()
}

// Path returns the path of the executable
func (c *Cmd) Path() string {
	return c.cmd.Path
}

// SetEnv appends the environment variables for the command
func (c *Cmd) SetEnv(env map[string]string) error {
	for k, v := range env {
		if !types.IsIdentifier(k) {
			return ErrBadParameter.Withf("Invalid environment variable name: %q", k)
		}
		c.cmd.Env = append(c.cmd.Env, fmt.Sprintf("%s=%q", k, v))
	}
	// return success
	return nil
}

// Return whether command has exited
func (c *Cmd) Exited() bool {
	if c.cmd.ProcessState == nil {
		return false
	} else {
		return c.cmd.ProcessState.Exited()
	}
}

// Return the pid of the process or 0
func (c *Cmd) Pid() int {
	if c.cmd.Process == nil {
		return 0
	} else {
		return c.cmd.Process.Pid
	}
}

// Send signal to the process
func (c *Cmd) Signal(s os.Signal) error {
	pid := c.Pid()
	if pid == 0 || c.Exited() {
		return ErrOutOfOrder.With("Cannot signal exited process")
	} else if err := syscall.Kill(pid, s.(syscall.Signal)); err != nil {
		return err
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t *Cmd) read(r io.ReadCloser, fn CallbackFn) {
	buf := bufio.NewReader(r)
	for {
		if line, err := buf.ReadBytes('\n'); err == io.EOF {
			return
		} else if err != nil && t.Err != nil {
			t.Err(t, []byte(err.Error()))
		} else {
			fn(t, line)
		}
	}
}
