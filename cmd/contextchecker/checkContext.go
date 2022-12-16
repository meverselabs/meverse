package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/meverselabs/meverse/cmd/app"
	"github.com/meverselabs/meverse/cmd/closer"
	"github.com/meverselabs/meverse/cmd/config"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/service/apiserver/zipcontext"
)

func testContext(exSt *chain.Store, cfg Config, cfgPath string) {
	ChainID := big.NewInt(0x1D5E)
	Version := uint16(0x0001)

	if err := setupContextFolder(&cfg); err != nil {
		panic(err)
	}

	ObserverKeys := []common.PublicKey{}
	for _, k := range cfg.ObserverKeys {
		pubkey, err := common.ParsePublicKey(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubkey)
	}
	cm := closer.NewManager()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cm.CloseAll()
	}()
	defer cm.CloseAll()

	//MaxBlocksPerFormulator := uint32(10)

	var InitGenesisHash hash.Hash256
	if len(cfg.CheckInitGenesisHash) > 0 {
		InitGenesisHash = hash.HexToHash(cfg.CheckInitGenesisHash)
	}
	var InitHash hash.Hash256
	if len(cfg.CheckInitHash) > 0 {
		InitHash = hash.HexToHash(cfg.CheckInitHash)
	}

	cdb, err := piledb.Open(cfg.CheckStoreRoot+"/chain", InitHash, cfg.CheckInitHeight, cfg.CheckInitTimestamp)
	if err != nil {
		panic(err)
	}
	cdb.SetSyncMode(true)
	st, err := chain.NewStore(cfg.CheckStoreRoot+"/context", cdb, ChainID, Version)
	if err != nil {
		panic(err)
	}
	cm.Add("store", st)

	cn := chain.NewChain(ObserverKeys, st, "")
	zipContext := zipcontext.NewZipContextService(nil, st, cfg.SaveZipPath, cfg.ZipInterval)
	cn.MustAddService(zipContext)

	if cfg.CheckInitHeight == 0 {
		if err := cn.UpdateInit(app.Genesis()); err != nil {
			panic(err)
		}
	} else {
		app.RegisterContractClass()
		if err := cn.InitWith(InitGenesisHash, InitHash, cfg.CheckInitHeight, cfg.CheckInitTimestamp); err != nil {
			panic(err)
		}
	}
	cm.RemoveAll()
	cm.Add("chain", cn)

	// initH := cfg.InitHeight
	// if initH == 0 {
	// 	initH = 1
	// }
	initH := st.Height() + 1

	for i := initH; i <= exSt.Height(); i++ {
		b, err := exSt.Block(i)
		if err != nil {
			panic(err)
		}
		if err := cn.UpdateBlock(b, nil); err != nil {
			panic(err)
		}
		if i%1000 == 0 {
			log.Println(i)
		}
	}

}

func setupContextFolder(cfg *Config) error {
	npath, err := filepath.Abs(cfg.StoreRoot)
	if err != nil {
		return err
	}
	cpath, err := filepath.Abs(cfg.CheckStoreRoot)
	if err != nil {
		return err
	}
	log.Println("npath, cpath", npath, cpath)
	if npath == cpath {
		return errors.New("invalid target path")
	}
	err = os.RemoveAll(cfg.CheckStoreRoot)
	if err != nil {
		return err
	}

	if cfg.ZipHeight > 1 {
		if err := unzip(fmt.Sprintf("%vmeverse_context_%v.zip", cfg.SaveZipPath, cfg.ZipHeight), fmt.Sprintf("./temp/%v", cfg.ZipHeight)); err != nil {
			return err
		}

		zipContextPath := fmt.Sprintf("./temp/%v/", cfg.ZipHeight)
		var zipCfg Config
		if err := config.LoadFile(zipContextPath+"_config.toml", &zipCfg); err != nil {
			return err
		}
		cfg.CheckInitGenesisHash = zipCfg.InitGenesisHash
		cfg.CheckInitHash = zipCfg.InitHash
		cfg.CheckInitHeight = zipCfg.InitHeight
		cfg.CheckInitTimestamp = zipCfg.InitTimestamp

		if err := os.Rename(zipContextPath+"_config.toml", zipContextPath+"data/_config.toml"); err != nil {
			return err
		}
		if err := os.Rename(zipContextPath+"data", cfg.CheckStoreRoot); err != nil {
			return err
		}
		err = os.RemoveAll(zipContextPath)
		if err != nil {
			return err
		}

	}

	return nil
}

func unzip(zipPath, dst string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return errors.New("invalid file path")
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	return nil
}
