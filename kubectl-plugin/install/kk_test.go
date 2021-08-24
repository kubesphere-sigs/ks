package install

import (
	"fmt"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVersionCheck(t *testing.T) {
	table := []struct {
		param     string
		expect    string
		returnErr bool
		message   string
	}{{
		param:     types.KsVersion,
		expect:    types.KsVersion,
		returnErr: false,
		message:   "do nothing with the default version",
	}, {
		param:     "nightly",
		expect:    fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -1).Format("20060102")),
		returnErr: false,
		message:   "this is the latest nightly version",
	}, {
		param:     "nightly-latest",
		expect:    fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -1).Format("20060102")),
		returnErr: false,
		message:   "this is the latest nightly version",
	}, {
		param:     "nightly-20060102",
		expect:    "nightly-20060102",
		returnErr: false,
		message:   "this is a specific nightly version",
	}, {
		param:     "fake",
		returnErr: true,
		message:   "this is a fake version",
	}, {
		param:     "nightly-fake",
		returnErr: true,
		message:   "this is a fake nightly version",
	}}

	for _, item := range table {
		opt := &kkOption{
			version: item.param,
		}

		err := opt.versionCheck()
		if item.returnErr {
			assert.NotNil(t, err, item.message)
		} else {
			assert.Nil(t, err, item.message)
			assert.Equal(t, item.expect, opt.version, item.message)
		}
	}
}

func TestSetDefaultIfNotExist(t *testing.T) {
	data := []byte("data")
	file := filepath.Join(os.TempDir(), "fake")
	defer func(filePath string) {
		_ = os.RemoveAll(filePath)
	}(file)

	err := setDefaultIfNotExist(data, file)
	assert.Nil(t, err, "failed to setDefaultIfNotExist")
	gotData, err := ioutil.ReadFile(file)
	assert.Nil(t, err, "failed to get the got data file")
	assert.Equal(t, string(data), string(gotData))
}
