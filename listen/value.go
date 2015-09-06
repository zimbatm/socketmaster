package listen

import (
	"fmt"
	"os"
	"strings"
)

// Used for command-line parsing to collect an array of listener files
type Value struct {
	addrs []string
	files []*os.File
}

func NewValue() *Value {
	return &Value{[]string{}, []*os.File{}}
}

func (self *Value) Set(addr string) error {
	file, _, err := Listen(addr)
	if err != nil {
		return fmt.Errorf("Unable to open %s: %s", addr, err.Error())
	}
	self.addrs = append(self.addrs, addr)
	self.files = append(self.files, file)
	return nil
}

func (self *Value) String() string {
	return strings.Join(self.addrs, ", ")
}

func (self *Value) Value() []*os.File {
	return self.files
}
