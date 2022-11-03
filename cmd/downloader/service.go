package main

import (
	"bytes"
	"encoding/json"
	"github.com/carfloresf/reddit-bot/internal/badger"
	"github.com/carfloresf/reddit-bot/internal/pushreddit"
	log "github.com/sirupsen/logrus"
)

type StoreService struct {
	db badger.DB
}

func NewStoreService(db badger.DB) *StoreService {
	return &StoreService{
		db: db,
	}
}

func (ss *StoreService) Store(subreddit string, receiveChan <-chan []pushreddit.Subreddit) {
	go func() {
		for posts := range receiveChan {
			log.Printf("received %d posts", len(posts))

			for _, post := range posts {
				postID := postPrefix + post.ID

				resultBool, err := ss.db.Has([]byte(subreddit), []byte(postID))
				if err != nil {
					log.Fatalf("badgerDB has failed %s", err)
				}

				post.ID = postID

				if !resultBool {
					reqBodyBytes := new(bytes.Buffer)
					err := json.NewEncoder(reqBodyBytes).Encode(post)
					if err != nil {
						log.Errorf("json encode failed %s", err)
						return
					}

					err = ss.db.Set([]byte(subreddit), []byte(post.ID), reqBodyBytes.Bytes())
					if err != nil {
						log.Errorf("badgerDB set failed %s", err)
						return
					}
				}

			}
		}
	}()
}
