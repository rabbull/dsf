package main

import (
	"context"
	"log"
	"time"

	"github.com/rabbull/dsf"
)

type LogLevel uint64

const (
	Debug LogLevel = iota
	Info
	Error
	None
)

type LoggerImpl struct {
	ID  int64
	Lvl LogLevel
}

func (lg *LoggerImpl) Context(ctx context.Context) dsf.Logger {
	return lg
}

func (lg *LoggerImpl) Debugf(fmt string, args ...interface{}) {
	if lg.Lvl > Debug {
		return
	}
	log.Printf("[D][%v][%v]"+fmt, append([]interface{}{
		lg.ID, time.Now().UnixNano(),
	}, args...)...)
}

func (lg *LoggerImpl) Logf(fmt string, args ...interface{}) {
	if lg.Lvl > Info {
		return
	}
	log.Printf("[I][%v][%v]"+fmt, append([]interface{}{
		lg.ID, time.Now().UnixNano(),
	}, args...)...)
}

func (lg *LoggerImpl) Errorf(fmt string, args ...interface{}) {
	if lg.Lvl > Error {
		return
	}
	log.Printf("[E][%v][%v]"+fmt, append([]interface{}{
		lg.ID, time.Now().UnixNano(),
	}, args...)...)
}
