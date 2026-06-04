package ipc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Action interface {
	ToString() string
	FromString([]string) error
}

var (
	WrongFieldNum = errors.New("Wrong amount of args fed to FromString")
	ParseError    = errors.New("Failed to parse")
)

type Mode int

const (
	Nothing Mode = iota
	Battery
)

type SetMode struct {
	Mode Mode
}

func (s *SetMode) ToString() string {
	return fmt.Sprintf("SetMode %d\n", s.Mode)
}

func (s *SetMode) FromString(args []string) error {

	if len(args) != 1 {
		return WrongFieldNum
	}

	mode, err := strconv.Atoi(args[0])

	if err != nil {
		return err
	}

	s.Mode = Mode(mode)

	return nil
}

func ParseMessage(message string) (Action, error) {
	var action Action
	var err error

	split := strings.Split(message, " ")

	switch split[0] {
	case "SetMode":
		action = &SetMode{}

		err = action.FromString(split[1:])

	default:
		err = ParseError
	}

	return action, err
}
