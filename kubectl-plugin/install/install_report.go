package install

import (
	"runtime"
	"time"
)

func (i *installReport) init() installReport {
	return installReport{
		os:        runtime.GOOS,
		arch:      runtime.GOARCH,
		beginTime: time.Now().String(),
	}
}

func (i *installReport) end() {
	i.endTime = time.Now().String()
}
