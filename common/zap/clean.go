package zap

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type clean struct {
	sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	dir            string
	retentionHours int64
}

func newClean(file string, retentionHours int) *clean {
	ctx, cancel := context.WithCancel(context.TODO())
	c := &clean{
		ctx:            ctx,
		cancel:         cancel,
		dir:            filepath.Dir(file),
		retentionHours: int64(retentionHours) * CleanCycle,
	}
	c.Add(1)
	go c.working()
	return c
}

func (c *clean) Close() {
	c.cancel()
	c.Wait()
}

func (c *clean) working() {
	defer c.Done()
	// 优先清理老历史文件
	cleanOnce(c)

	sec := cleanTime() + CleanCycle - time.Now().Unix()

	t := time.NewTimer(time.Duration(sec) * time.Second)
	defer t.Stop()
loop:
	for {
		select {
		case <-c.ctx.Done():
			break loop
		case <-t.C:
		}
		t.Reset(time.Duration(CleanCycle) * time.Second)
		cleanOnce(c)
	}
}

func cleanTime() int64 {
	t, _ := time.ParseInLocation(CleanFormat, time.Now().Format(CleanFormat), time.Local)
	return t.Unix()
}

func cleanOnce(c *clean) {
	files, _ := ioutil.ReadDir(c.dir)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		fileTime, err := time.ParseInLocation(ExtFormat, ext, time.Local)
		if err != nil {
			continue
		}
		if fileTime.Unix() < (cleanTime() - c.retentionHours) {
			os.Remove(fmt.Sprintf("%s/%s", c.dir, file.Name()))
		}
	}
}
