package common

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
)

var ErrPidfileExists = errors.New("pidfile already exists")
var ErrPidfileDoesntExist = errors.New("pidfile doesn't exist")

func GetPid(pidfile string) (int, error) {
	if !exists(pidfile) {
		return 0, ErrPidfileDoesntExist
	}

	data, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}

func InitPidfile(pidfile string) error {
	if exists(pidfile) {
		return ErrPidfileExists
	}

	pid := string(os.Getpid())
	return ioutil.WriteFile(pidfile, []byte(pid), 0644)
}

func RemovePidfile(pidfile string) error {
	return os.Remove(pidfile)
}


func exists(f string) bool {
	_, err := os.Stat(f)
	return !os.IsNotExist(err)
}