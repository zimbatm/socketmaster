package process

import (
	"log"
	"os"
	"syscall"
)

// Manager states
const (
	MANAGER_STARTING = ManagerState(iota)
	MANAGER_RUNNING
	MANAGER_STOPPING
)

type ManagerState int

func (x ManagerState) String() string {
	switch x {
	case MANAGER_STARTING:
		return "starting"
	case MANAGER_RUNNING:
		return "running"
	case MANAGER_STOPPING:
		return "stopping"
	}
	return "BUG: unknown state"
}

type Manager struct {
	done chan error
}

func StartManager(c *Config, logger *log.Logger, signals chan os.Signal, managerReady chan<- bool) (m *Manager) {
	m = &Manager{make(chan error, 1)}
	m.run(c, logger, signals, managerReady)
	return
}

func (m *Manager) run(c *Config, logger *log.Logger, signals chan os.Signal, managerReady chan<- bool) {
	var (
		current   *Process
		err       error
		events    chan Event   = make(chan Event)
		readyChan chan int     = make(chan int)
		set       Set          = NewSet()
		state     ManagerState = MANAGER_STARTING
	)
	defer func() {
		// When the manager stops, everything has to die with it
		set.KillAll()
		m.done <- err
	}()

	startProcess := func() (err error) {
		current, err = StartProcess(c, logger, events)
		if err == nil {
			set.Add(current)
		}
		return
	}

	if err = startProcess(); err != nil {
		return
	}

	if c.NotifyConn != nil {
		// TODO: ready messages from the socket and send ready messages
	}

	changeState := func(newState ManagerState) {
		logger.Printf("manager={state=%s}", newState)

		switch newState {
		case MANAGER_RUNNING:
			managerReady <- true
		case MANAGER_STOPPING:
			if current != nil {
				current.Shutdown()
				current = nil
			}
		}

		state = newState
	}

	for {
		if state == MANAGER_STOPPING {
			if set.Empty() {
				logger.Printf("manager={state=%s} msg=Bye", state)
				return
			} else {
				logger.Printf("manager={state=%s} remaining=%v", state, set.ToList())
			}
		}

		select {
		case signal := <-signals:
			logger.Printf("manager={state=%s} signal=%v", state, signal)
			sig := signal.(syscall.Signal)
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				if state < MANAGER_STOPPING {
					changeState(MANAGER_STOPPING)
				}
			case syscall.SIGHUP:
				// Reload
				if state == MANAGER_RUNNING {
					if err = startProcess(); err != nil {
						changeState(MANAGER_STOPPING)
					}
				} else {
					logger.Println("manager={state=%s} msg='Cannot reload'", state)
				}
			case syscall.SIGUSR1, syscall.SIGUSR2:
				// Forward signals to the current process
				if current != nil {
					current.Signal(signal)
				}
			default:
				// Just ignore
			}
		case pid := <-readyChan:
			if current != nil && current.Pid == pid {
				current.NotifyReady()
			}
		case e := <-events:
			logger.Printf("manager={state=%s} event={%v}", state, e)

			switch e.State {
			case PROCESS_READY:
				if state < MANAGER_RUNNING {
					changeState(MANAGER_RUNNING)
				}
				if e.Process == current {
					set.Each(func(p *Process) {
						// There should be at most one here
						if p != current {
							p.Shutdown()
						}
					})
				} else {
					logger.Printf("manager={state=%s} msg='Unexpected event'")
				}
			case PROCESS_STOPPING:
				if current == e.Process {
					changeState(MANAGER_STOPPING)
				}
			case PROCESS_DIED:
				set.Remove(e.Process)
				if current == e.Process {
					current = nil
					changeState(MANAGER_STOPPING)
				}
			default:
				logger.Fatalf("manager={state=%s} msg='Unexpected event' event={%v}", state, e)
			}
		}
	}
}

func (m *Manager) Wait() (err error) {
	return <-m.done
}
