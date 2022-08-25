package main

import (
	"github.com/dgraph-io/badger/v3"

	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var BadgerDB *badger.DB

func Open(path string) (err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false
	opts.ValueThreshold = 256
	opts.CompactL0OnClose = true
	BadgerDB, err = badger.Open(opts)
	if err != nil {
		log.Println("badger open failed", "path", path, "err", err)
		return err
	}
	return nil
}

func Close() {
	err := BadgerDB.Close()
	if err == nil {
		log.Errorf("closed badger db")
	} else {
		log.Errorf("failed to close badger db: %s", err)
	}
}

func Set(key []byte, value []byte) (err error) {
	wb := BadgerDB.NewWriteBatch()
	defer wb.Cancel()
	err = wb.SetEntry(badger.NewEntry(key, value).WithMeta(0))
	if err != nil {
		log.Println("Failed to write data to cache.", "key", string(key), "value", string(value), "err", err)
		return err
	}
	err = wb.Flush()
	if err != nil {
		log.Println("Failed to flush data to cache.", "key", string(key), "value", string(value), "err", err)
	}

	return nil
}

func Get(key []byte) (value string, err error) {
	var ival []byte
	err = BadgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		ival, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Println("Failed to read data from the cache.", "key", string(key), "error", err)
		return "", err
	}
	return string(ival), nil
}

func Has(key []byte) (bool, error) {
	var exist bool = false
	err := BadgerDB.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return err
		} else {
			exist = true
		}
		return err
	})

	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			err = nil
		}
	}

	return exist, err
}
