package zap

import (
	"os"
	"time"
)

type writer struct {
	file     string
	ext      string
	lastfile string
	f        *os.File
}

func newWriter(file string) *writer {
	return &writer{
		file: file,
		ext:  ExtFormat,
	}
}

func (w *writer) Write(b []byte) (int, error) {
	nowfile := w.file + time.Now().Format(w.ext)
	if nowfile != w.lastfile {
		if w.f != nil {
			if err := w.close(); err != nil {
				return 0, err
			}
		}
		f, err := os.OpenFile(nowfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}
		w.f = f
		w.lastfile = nowfile
	}
	return w.f.Write(b)
}

func (w *writer) Sync() error {
	return w.f.Sync()
}

func (w *writer) close() error {
	if w.f == nil {
		return nil
	}
	if err := w.f.Sync(); err != nil {
		return err
	}
	if err := w.f.Close(); err != nil {
		return err
	}
	w.f = nil
	w.lastfile = ""
	return nil
}
