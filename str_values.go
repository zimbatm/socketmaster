package main

import (
	"strings"
)

// Allows to accumulate a list of strings for command-line parsing
type StrValues struct {
	strs []string
}

func NewStrValues() *StrValues {
	return &StrValues{[]string{}}
}

func (self *StrValues) Set(str string) error {
	self.strs = append(self.strs, str)
	return nil
}

func (self *StrValues) String() string {
	return strings.Join(self.strs, ", ")
}

func (self *StrValues) Value() []string {
	return self.strs
}
