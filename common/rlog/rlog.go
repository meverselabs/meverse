package rlog

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fletaio/fleta/common/binutil"
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

var logger *log.Logger
var loggerAddress string
var loggerHost string
var loggerUse bool

func init() {
	logger = log.New(os.Stderr, "", log.LstdFlags)
}

// Println calls l.Output to print to the logger.
func Println(v ...interface{}) {
	logger.Println(v...)
}

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	logger.Fatalln(v...)
}

// errors
var (
	ErrNoLog = errors.New("no log")
)

// SetRLogAddress sets the addres of the current node
func SetRLogAddress(address string) {
	loggerAddress = address
}

// SetRLogHost sets the host of the remote log
func SetRLogHost(host string) {
	loggerHost = host
}

// Enablelogger turn on of the remote log
func Enablelogger(path string) {
	loggerUse = true

	lw, err := NewLogWriter(path)
	if err != nil {
		panic(err)
	}
	logger = log.New(lw, "", log.LstdFlags)

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
	if !loggerUse {
		return len(bs), nil
	}

	if bs[len(bs)-1] == '\n' {
		bs = bs[:len(bs)-1]
	}
	if len(bs) > 65535 {
		bs = bs[:65535]
	}

	var buffer bytes.Buffer
	buffer.Write(binutil.LittleEndian.Uint64ToBytes(uint64(time.Now().UnixNano())))
	buffer.Write(binutil.LittleEndian.Uint16ToBytes(uint16(len(bs))))
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

// Upload uploads logs to the remote log server
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
		buffer.Write(bs)
		appended++
	}
	if buffer.Len() == 0 {
		return ErrNoLog
	}
	req, err := http.Post(loggerHost+"/api/addresses/"+loggerAddress+"/logs", "application/json", bytes.NewReader(buffer.Bytes()))
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
