package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	reddit "github.com/vartanbeno/go-reddit/v2/reddit"

	"github.com/carfloresf/reddit-bot/config"
	badger "github.com/carfloresf/reddit-bot/internal/badger"
	"github.com/carfloresf/reddit-bot/internal/pushreddit"
)

const (
	subreddit   = "overemployed"
	indexPrefix = "index_"
	indexKey    = indexPrefix + subreddit

	postPrefix = "post_"

	find  = true
	total = 100000000

	count = true
)

func main() {
	cfg, err := config.ReadConfig("config/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	badgerDB, err := badger.NewBadgerDB(cfg.DB.DBFile)
	if err != nil {
		log.Fatalf("badgerDB open failed %s", err)
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			log.Fatalf("badgerDB close failed %s", err)
		}
	}()

	index := 0
	indexVal, err := badgerDB.Get([]byte(indexPrefix), []byte(indexKey))
	if err != nil {
		log.Errorf("badgerDB get failed %s", err)
	} else {
		index = cast.ToInt(string(indexVal))
	}

	log.Infof("index %d", index)

	// add your reddit username and password, secret and client id in here
	credentials := reddit.Credentials{ID: cfg.Reddit.ClientID, Secret: cfg.Reddit.Secret, Username: cfg.Reddit.Username, Password: cfg.Reddit.Password}
	client, err := reddit.NewClient(credentials)
	if err != nil {
		log.Fatalf("reddit client failed %s", err)
	}

	_, _, err = client.Subreddit.Get(context.Background(), subreddit)
	if err != nil {
		log.Fatalf("reddit get subreddit failed %s", err)
	}

	clientPush := pushreddit.NewClient()

	now := time.Now()

	keysp, err := badgerDB.IterateKeys()
	if err != nil {
		log.Fatalf("badgerDB iterate failed %s", err)
	}

	log.Printf("keys: %d", len(keysp))

	var receiveChan = make(chan []pushreddit.Subreddit, 10000)

	storeService := NewStoreService(badgerDB)
	storeService.Store(subreddit, receiveChan)

	go func() {
		for i := index; i < total; i++ {
			posts, err := clientPush.GetPostsSubreddit(subreddit, now.Add(time.Duration(-i)*time.Hour*6), now.Add(time.Duration(-(i-1))*time.Hour*6), 100000000)
			if err != nil {
				log.Fatalf("reddit get subreddit posts failed %s", err)
			}

			receiveChan <- posts.Data

			log.Println("index", i)
			err = badgerDB.Set([]byte(indexPrefix), []byte(indexKey), []byte(cast.ToString(i)))
			if err != nil {
				log.Fatalf("badgerDB set failed %s", err)
			}
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	log.Infof("waiting for store service to finish")

	for {
		if len(receiveChan) == 0 {
			close(receiveChan)
			break
		}
	}

	log.Info("shutting down...")
}
