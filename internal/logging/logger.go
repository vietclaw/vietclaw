package logging

import (
	"io"
	"log"
	"os"
)

func New(path string) (*log.Logger, *os.File, error) {
	file, _, err := EnsureLogFile(path)
	if err != nil {
		return nil, nil, err
	}

	writer := io.MultiWriter(os.Stdout, file)
	logger := log.New(writer, "vietclaw ", log.LstdFlags)
	return logger, file, nil
}
