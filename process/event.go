package process

import "fmt"

type Event struct {
	*Process
	State ProcessState
	Err   error
}

func (e Event) String() string {
	if e.Err == nil {
		return fmt.Sprintf("%v state=%v", e.Process, e.State)
	} else {
		return fmt.Sprintf("%v state=%v err=%v", e.Process, e.State, e.Err)
	}
}
