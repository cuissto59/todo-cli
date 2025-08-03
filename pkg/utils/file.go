package utils

import (
	"fmt"
	"os"
	"syscall"
)

const FilePath = "db/todo.csv"

func LoadFile(filepath string) (*os.File, error) {
	f, err := os.OpenFile(
		filepath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		os.ModePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file for reading")
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err

	}

	return f, nil
}

func CloseFile(f *os.File) error {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return f.Close()
}
