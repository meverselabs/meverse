package zipcontext

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/keydb"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/apiserver"
)

// ZipContextService is provides initContext info files
type ZipContextService struct {
	types.ServiceBase
	api         *apiserver.APIServer
	st          *chain.Store
	savePath    string
	zipInterval uint32
}

// NewZipContextService returns a ZipContextService
func NewZipContextService(api *apiserver.APIServer, st *chain.Store, savePath string, zipInterval uint32) *ZipContextService {
	if strings.LastIndex(savePath, "/") != len(savePath)-1 {
		savePath = savePath + "/"
	}

	s := &ZipContextService{
		api:         api,
		st:          st,
		savePath:    savePath,
		zipInterval: zipInterval,
	}
	api.AddGETPath("zipcontext", s.zipContext)
	api.AddGETPath("makeZipcontext", s.makeZipContext)
	return s
}

// Name returns the name of the service
func (s *ZipContextService) Name() string {
	return "meverse.zipContextService"
}

// OnBlockConnected called when a block is connected to the chain
func (s *ZipContextService) OnBlockConnected(b *types.Block, loader types.Loader) {
	if b.Header.Height%s.zipInterval == 0 {
		// savePath := "./zipcontext/"
		s.ZipContext(s.savePath)
	}
}

func (s ZipContextService) zipContext(c echo.Context) error {
	files, err := ioutil.ReadDir(s.savePath)
	if err != nil {
		return err
	}

	heightStr := strings.ToLower(c.QueryParam("height"))
	var filePath string
	if heightStr == "" {
		var height uint64
		for _, file := range files {
			fName := file.Name()
			if strings.Contains(fName, "meverse_context_") && strings.Contains(fName, ".zip") {
				HeightStr := strings.TrimLeft(fName, "mevrs_contx")
				HeightStr = strings.TrimRight(HeightStr, ".zip")
				h, err := strconv.ParseUint(HeightStr, 10, 64)
				if err != nil {
					fmt.Printf("%+v", err)
					continue
				}
				if height < h {
					height = h
					filePath = s.savePath + fName
				}
			}
		}
	} else {
		_, err := strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			return err
		}
		filePath = "meverse_context_" + heightStr + ".zip"

	}
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return errors.New("not exist zipcontext")
	}
	return c.File(filePath)

}

type initContextInfo struct {
	Height     uint32
	TargetHash hash.Hash256
	Timestamp  uint64
	GenHash    hash.Hash256
}

func (ici *initContextInfo) zipContextInfo(filePath string, zipContextPath string) error {
	err := os.MkdirAll(filepath.Dir(filePath), os.ModeDir)
	if err != nil {
		return err
	}

	archive, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	contextFile, err := os.Open(zipContextPath)
	if err != nil {
		return err
	}
	defer contextFile.Close()

	w1, err := zipWriter.Create("data/context")
	if err != nil {
		return err
	}
	if _, err := io.Copy(w1, contextFile); err != nil {
		return err
	}

	w2, err := zipWriter.Create("_config.toml")
	if err != nil {
		return err
	}
	fmt.Fprintf(w2, "InitGenesisHash = \"%v\"\n", ici.GenHash.String())
	fmt.Fprintf(w2, "InitHeight = %v\n", ici.Height)
	fmt.Fprintf(w2, "InitHash = \"%v\"\n", ici.TargetHash.String())
	fmt.Fprintf(w2, "InitTimestamp = %v\n", ici.Timestamp)

	if err = zipWriter.Close(); err != nil {
		return err
	}
	return nil
}

func (s *ZipContextService) makeZipContext(c echo.Context) error {
	_, err := s.ZipContext(s.savePath)
	return err
}

func (s *ZipContextService) ZipContext(savePath string) (filePath string, err error) {
	if strings.LastIndex(savePath, "/") == len(savePath)-1 {
		savePath = savePath + "/"
	}

	zipContextPath := "./tempZipContext/context"
	if err = os.MkdirAll(filepath.Dir(zipContextPath), os.ModeDir); err != nil {
		return
	} else {
		defer func() {
			err = os.RemoveAll(filepath.Dir(zipContextPath))
		}()
	}

	if err = s.st.CopyContext(zipContextPath); err != nil {
		return
	}

	initContextInfo, err := s.extractInitContextInfo(zipContextPath)
	if err != nil {
		return
	}
	filePath = fmt.Sprintf("%vmeverse_context_%v.zip", savePath, initContextInfo.Height)
	if err = initContextInfo.zipContextInfo(filePath, zipContextPath); err != nil {
		return
	}

	return filePath, nil
}

var (
	tagHeight     = byte(0x01)
	tagHeightHash = byte(0x02)
)

func toHeightHashKey(height uint32) []byte {
	bs := make([]byte, 5)
	bs[0] = tagHeightHash
	bin.PutUint32(bs[1:], height)
	return bs
}

func (s *ZipContextService) extractInitContextInfo(contextPath string) (ici initContextInfo, err error) {
	db, err := keydb.Open(contextPath, func(key []byte, value []byte) (interface{}, error) {
		switch key[0] {
		case tagHeight:
			return bin.Uint32(value), nil
		case tagHeightHash:
			var h hash.Hash256
			h.SetBytes(value)
			return h, nil
		}
		return nil, nil
	})
	if err != nil {
		return
	}
	defer db.Close()
	if err = db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get([]byte{tagHeight})
		if err != nil {
			return err
		}
		ici.Height = value.(uint32)
		return nil
	}); err != nil {
		return
	}

	if ici.Height <= 1 {
		err = errors.New("height must be greater than 1")
		return
	}

	if err = db.View(func(txn *keydb.Tx) error {
		value, err := txn.Get(toHeightHashKey(0))
		if err != nil {
			return err
		}
		var ok bool
		ici.GenHash, ok = value.(hash.Hash256)
		if !ok {
			return errors.New("not found genhash")
		}
		return nil
	}); err != nil {
		return
	}
	b, err := s.st.Block(ici.Height)
	if err != nil {
		return
	}
	ici.Timestamp = b.Header.Timestamp

	bsHeader, _, err := bin.WriterToBytes(&b.Header)
	if err != nil {
		return
	}
	ici.TargetHash = hash.Hash(bsHeader)

	return
}
