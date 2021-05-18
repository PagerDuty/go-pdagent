package common

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

var ErrPidfileExists = errors.New("pidfile already exists")
var ErrPidfileDoesntExist = errors.New("pidfile doesn't exist")

func IsRunning(pidfile string) (bool, error) {
	proc, err := getProcess(pidfile)
	if err != nil {
		return false, err
	}

	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetPid(pidfile string) (int, error) {
	exists, err := fileExists(pidfile)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, ErrPidfileDoesntExist
	}

	data, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}

func InitPidfile(pidfile string) error {
	exists, err := fileExists(pidfile)
	if err != nil {
		return err
	}

	if exists {
		return ErrPidfileExists
	}

	pid := strconv.Itoa(os.Getpid())
	return ioutil.WriteFile(pidfile, []byte(pid), 0744)
}

func RemovePidfile(pidfile string) error {
	return os.Remove(pidfile)
}

func TerminateProcess(pidfile string) error {
	proc, err := getProcess(pidfile)
	if err != nil {
		return err
	}

	return proc.Signal(syscall.SIGTERM)
}

func getProcess(pidfile string) (*os.Process, error) {
	pid, err := GetPid(pidfile)
	if err != nil {
		return nil, err
	}

	return os.FindProcess(pid)
}
