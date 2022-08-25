package main

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func main() {

	err := Open("tmp/badger")
	if err != nil {
		log.Fatalf("badger open failed %s", err)
	}
	defer Close()

	// add your reddit username and password, secret and client id in here
	credentials := reddit.Credentials{ID: "-", Secret: "-", Username: "-", Password: "-"}
	client, err := reddit.NewClient(credentials)
	if err != nil {
		log.Fatalf("reddit client failed %s", err)
	}

	comments, _, err := client.User.Comments(context.Background(), &reddit.ListUserOverviewOptions{
		ListOptions: reddit.ListOptions{
			Limit: 100,
		},
		Time: "all",
	})

	if err != nil {
		log.Fatalln(err)
	}

	for _, comment := range comments {
		if time.Now().Sub(comment.Created.Time) > time.Hour*24*7 {
			log.Printf("%s:: %s // %s %s \n", comment.ID, comment.Body, comment.Created, comment.PostPermalink)

			exists, err := Has([]byte(comment.ID))
			if err != nil {
				log.Errorf("failed to check if comment exists %s", err)
				break
			}
			if !exists {
				time.Sleep(time.Second * 10)

				comment, response, err := client.Comment.Edit(context.Background(), "t1_"+comment.ID, "comment edited")
				if err != nil {
					log.Error("fatal error ", err)
					break
				}

				log.Printf("%+v\n", response)

				err = Set([]byte(comment.ID), []byte(comment.Body))
				if err != nil {
					log.Error("fatal error ", err)
					break
				}
			} else {
				log.Println("comment already exists", comment.ID)
			}
		}
	}
}
