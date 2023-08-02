package main

import (
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/unix"
)

const (
	appPath = "/tmp/enclave-application"
)

type transparencyLog interface {
	append(*logRecord) error
	String() string // human-readable representation
}

type appRetriever interface {
	retrieve(*Enclave, chan []byte) error
}

type appLoader struct {
	enclave   *Enclave
	log       transparencyLog
	app       chan []byte
	appExited chan struct{}
	appRetriever
}

func newAppLoader(e *Enclave, r appRetriever) *appLoader {
	return &appLoader{
		enclave:      e,
		app:          make(chan []byte),
		appExited:    make(chan struct{}),
		appRetriever: r,
		log:          &memLog{},
	}
}

func writeToDisk(appBlob []byte) error {
	_ = unix.Umask(0)
	return os.WriteFile(appPath, appBlob, 0755)
}

func (l *appLoader) startCmd() {
	cmd := exec.Command(appPath)
	err := cmd.Run()
	elog.Printf("Enclave application exited with error: %v", err)
	l.appExited <- struct{}{}
}

func (l *appLoader) appendToLog(app []byte) {
	l.log.append(newLogRecord(app))
}

func (l *appLoader) runApp() chan error {
	e := make(chan error)
	var err error

	go l.retrieve(l.enclave, l.app)
	go func() {
		for {
			select {
			case <-l.appExited:
				time.Sleep(time.Second)
				elog.Println(l.log)
				go l.startCmd()

			case app := <-l.app:
				if err = writeToDisk(app); err != nil {
					e <- err
					return
				}
				l.appendToLog(app)
				go l.startCmd()
			}
		}
	}()

	return e
}
