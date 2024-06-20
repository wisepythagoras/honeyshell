package core

import (
	"log"
	"os"
	"sync"
)

// Logman is a struct that handles logging to a file.
type Logman struct {
	filename string
	*log.Logger
}

var theLog *Logman
var once sync.Once

// GetLogmanInstance gets (or create) the instance of the logger.
func GetLogmanInstance() *Logman {
	once.Do(func() {
		theLog = CreateLogmanLogger("honeyshell.log")
	})

	return theLog
}

// CreateLogmanLogger creates a logger.
func CreateLogmanLogger(fname string) *Logman {
	file, _ := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

	return &Logman{
		filename: fname,
		Logger:   log.New(file, "", log.Ldate|log.Ltime),
	}
}
