package process

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	PROCESS_STARTING = ProcessState(iota)
	PROCESS_READY
	PROCESS_STOPPING
	PROCESS_DIED
)

type ProcessState int

func (x ProcessState) String() string {
	switch x {
	case PROCESS_STARTING:
		return "starting"
	case PROCESS_READY:
		return "ready"
	case PROCESS_STOPPING:
		return "stopping"
	case PROCESS_DIED:
		return "died"
	}
	return "BUG: unknown state"
}

type Process struct {
	*os.Process
	ready    chan<- bool
	shutdown chan<- bool
}

func StartProcess(c *Config, logger *log.Logger, events chan<- Event) (p *Process, err error) {
	ioReader, ioWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	never := make(chan time.Time)
	var (
		startTimeout, stopTimeout <-chan time.Time = never, never
	)
	ready := make(chan bool, 100)
	shutdown := make(chan bool, 100) // External event, make some room
	died := make(chan error)

	// TODO: replace os.Stdin with /dev/null
	ps, err := createProcess(c, os.Stdin, ioWriter, ioWriter)
	if err != nil {
		return
	}
	p = &Process{ps, ready, shutdown}

	// Prefix lines with the process' pid
	go func(p *Process, r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			logger.Printf("[%d] %s", p.Pid, scanner.Text())
		}
	}(p, ioReader)

	// Handle the process death
	go func(p *Process, r io.ReadCloser, died chan<- error) {
		defer r.Close()
		for {
			state, err := p.Wait()
			if err != nil || state.Exited() {
				died <- err
				return
			}
		}
	}(p, ioReader, died)

	// Linearize and control the events
	go func() {
		var (
			err   error        = nil
			state ProcessState = PROCESS_STARTING
		)

		logger.Printf("process={%v state=%s}", p, state)

		// change state and emits back to the manager
		emit := func(newState ProcessState, newErr error) {
			if newErr == nil {
				logger.Printf("process={%v state=%s} new_state=%s", p, state, newState)
			} else {
				logger.Printf("process={%v state=%s} new_state=%s err=%v", p, state, newState, newErr)
			}

			state = newState
			err = newErr
			if newState == PROCESS_STOPPING {
				p.Kill()
			}
			events <- Event{p, newState, newErr}
		}

		if c.StartTimeout > 0 {
			startTimeout = time.After(c.StartTimeout)
		} else if c.NotifyConn == nil {
			emit(PROCESS_READY, nil)
		}

		for {
			select {
			case <-shutdown:
				if state < PROCESS_DIED {
					if c.StopTimeout > 0 {
						stopTimeout = time.After(c.StopTimeout)
					}
					emit(PROCESS_STOPPING, p.Signal(syscall.SIGTERM))
				}
			case <-ready:
				if state < PROCESS_STOPPING {
					emit(PROCESS_READY, err)
				}
			case err = <-died:
				logger.Printf("process={%v state=%s} msg='Process died' err=%v", p, state, err)
				emit(PROCESS_DIED, err)
				return
			case <-startTimeout:
				if state < PROCESS_READY {
					if c.NotifyConn == nil {
						emit(PROCESS_READY, err)
					} else {
						emit(PROCESS_STOPPING, fmt.Errorf("Didn't start in time"))
					}
				}
			case <-stopTimeout:
				if state < PROCESS_STOPPING {
					emit(PROCESS_DIED, fmt.Errorf("Didn't stop in time"))
				}
			}
		}
	}()

	return
}

func (p *Process) String() string {
	return fmt.Sprintf("pid=%d", p.Pid)
}

func (p *Process) NotifyReady() {
	p.ready <- true
}

func (p *Process) Shutdown() {
	p.shutdown <- true
}

func createProcess(c *Config, stdin, stdout, stderr *os.File) (*os.Process, error) {
	env := os.Environ()
	files := []*os.File{stdin, stdout, stderr}
	command := c.Command
	args := []string{c.Command}

	if len(c.Files) > 0 {
		// systemd socket activation, LISTEN_PID below
		env = append(env, fmt.Sprintf("LISTEN_FDS=%d", len(c.Files)))
		files = append(files, c.Files...)
		command = "/bin/sh"
		// working around the lack of fork+exec
		args = []string{"/bin/sh", "-c", fmt.Sprintf("LISTEN_PID=$$ exec %s", c.Command)}
	}

	if c.NotifyConn != nil {
		env = append(env,
			fmt.Sprintf("NOTIFY_SOCKET=%s", c.NotifyConn.LocalAddr().String()))
	}

	procAttr := &os.ProcAttr{
		Dir:   c.Dir,
		Env:   env,
		Files: files,
		Sys:   &syscall.SysProcAttr{
		// Pdeathsig: syscall.SIGKILL, TODO: Linux only
		},
	}

	if c.User != nil {
		uid, _ := strconv.Atoi(c.User.Uid)
		gid, _ := strconv.Atoi(c.User.Gid)

		procAttr.Sys.Credential = &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
	}

	return os.StartProcess(command, args, procAttr)
}
