package main

import (
	"context"
	"github.com/carfloresf/reddit-bot/config"
	"github.com/carfloresf/reddit-bot/internal/badger"
	log "github.com/sirupsen/logrus"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"time"
)

func main() {
	cfg, err := config.ReadConfig("config/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("config %+v", cfg)

	badgerDB, err := badger.NewBadgerDB(cfg.DB.DBFile)
	if err != nil {
		log.Fatalf("badgerDB open failed %s", err)
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			log.Fatalf("badgerDB close failed %s", err)
		}
	}()

	// add your reddit username and password, secret and client id in here
	credentials := reddit.Credentials{ID: cfg.Reddit.ClientID, Secret: cfg.Reddit.Secret, Username: cfg.Reddit.Username, Password: cfg.Reddit.Password}
	client, err := reddit.NewClient(credentials)
	if err != nil {
		log.Fatalf("reddit client failed %s", err)
	}

	comments, _, err := client.User.Comments(context.Background(), &reddit.ListUserOverviewOptions{
		ListOptions: reddit.ListOptions{
			Limit: 1000,
		},
		Time: "all",
	})
	if err != nil {
		log.Fatalln(err)
	}

	for _, comment := range comments {
		log.Infof("comment %s", comment.Body)
		if time.Now().Sub(comment.Created.Time) > time.Hour*24 {
			log.Printf("%s:: %s // %s %s \n", comment.ID, comment.Body, comment.Created, comment.PostPermalink)

			exists, err := badgerDB.Has([]byte("deleter"), []byte(comment.ID))
			if err != nil {
				log.Errorf("failed to check if comment exists %s", err)
				break
			}
			if !exists {
				time.Sleep(time.Second * 10)

				response, err := client.Comment.Delete(context.Background(), "t1_"+comment.ID)
				if err != nil {
					log.Error("fatal error ", err)
					break
				}

				log.Printf("%+v\n", response)

				err = badgerDB.Set([]byte("deleter"), []byte(comment.ID), []byte("deleted"))
				if err != nil {
					log.Error("fatal error ", err)
					break
				}
			} else {
				log.Println("comment already deleted", comment.ID)
			}
		}
	}
}
