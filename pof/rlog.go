package pof

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fletaio/fleta/common/util"

	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

var rlog *log.Logger
var rlogAddress string
var rlogHost string

var (
	ErrNoLog = errors.New("no log")
)

func setRLogAddress(address string) {
	rlogAddress = address
}

func setRLogHost(host string) {
	rlogHost = host
}

func init() {
	lw, err := NewLogWriter("./_rlog")
	if err != nil {
		panic(err)
	}
	rlog = log.New(lw, "", log.LstdFlags)

	go func() {
		for {
			if err := lw.Upload(); err != nil {
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

type LogWriter struct {
	sync.Mutex
	db *ledis.DB
}

func NewLogWriter(path string) (*LogWriter, error) {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = path
	l, err := ledis.Open(cfg)
	if err != nil {
		return nil, err
	}
	db, err := l.Select(0)
	if err != nil {
		return nil, err
	}
	lw := &LogWriter{
		db: db,
	}
	return lw, nil
}

func (lw *LogWriter) Write(bs []byte) (int, error) {
	os.Stderr.Write(bs)

	if bs[len(bs)-1] == '\n' {
		bs = bs[:len(bs)-1]
	}

	var buffer bytes.Buffer
	buffer.Write(util.Uint64ToBytes(uint64(time.Now().UnixNano())))
	buffer.Write(bs)

	lw.Lock()
	defer lw.Unlock()
	count, err := lw.db.LLen([]byte("log"))
	if err != nil {
		return 0, err
	}
	if count > 5000000 {
		lw.db.LPop([]byte("log"))
	}
	if _, err := lw.db.RPush([]byte("log"), buffer.Bytes()); err != nil {
		return 0, err
	}
	return len(bs), nil
}

func (lw *LogWriter) Upload() error {
	lw.Lock()
	count, err := lw.db.LLen([]byte("log"))
	lw.Unlock()
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	appended := int32(0)
	for i := int32(0); i < int32(count) && i < 100; i++ {
		bs, err := lw.db.LIndex([]byte("log"), i)
		if err != nil {
			break
		}
		item := &LogData{
			Timestamp: util.BytesToUint64(bs[:8]),
			Data:      string(bs[8:]),
		}
		body, err := json.Marshal(item)
		if err != nil {
			break
		}
		buffer.Write(body)
		appended++
	}
	if buffer.Len() == 0 {
		return ErrNoLog
	}
	req, err := http.Post(rlogHost+"/api/addresses/"+rlogAddress+"/logs", "application/json", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		return err
	}
	if req.StatusCode == 200 {
		lw.Lock()
		defer lw.Unlock()
		count, err := lw.db.LLen([]byte("log"))
		if err != nil {
			return err
		}
		if appended == int32(count) {
			if _, err := lw.db.LClear([]byte("log")); err != nil {
				return err
			}
		} else {
			if err := lw.db.LTrim([]byte("log"), int64(appended), count); err != nil {
				return err
			}
		}
	} else {
		return ErrNoLog
	}
	return nil
}

type LogData struct {
	Timestamp uint64
	Data      string
}
