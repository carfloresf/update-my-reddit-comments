package main

import (
	"context"
	"time"

	"github.com/carfloresf/reddit-bot/config"
	"github.com/carfloresf/reddit-bot/internal/badger"
	log "github.com/sirupsen/logrus"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

// isCommentDeleted checks if a comment is already marked as deleted in the DB.
func isCommentDeleted(db badger.DB, commentID string) (bool, error) {
	return db.Has([]byte("deleter"), []byte(commentID))
}

// markCommentAsDeleted marks the comment as deleted in the DB.
func markCommentAsDeleted(db badger.DB, commentID string) error {
	return db.Set([]byte("deleter"), []byte(commentID), []byte("deleted"))
}

// processComment handles the processing and deletion of a single Reddit comment.
func processComment(ctx context.Context, comment *reddit.Comment, client *reddit.Client, db badger.DB) {
	if time.Since(comment.Created.Time) > 24*time.Hour {
		log.WithFields(log.Fields{
			"commentID":     comment.ID,
			"body":          comment.Body,
			"created":       comment.Created,
			"postPermalink": comment.PostPermalink,
		}).Info("Comment eligible for deletion")

		exists, err := isCommentDeleted(db, comment.ID)
		if err != nil {
			log.WithFields(log.Fields{
				"commentID": comment.ID,
				"error":     err,
			}).Error("Failed to check if comment exists")
			return
		}
		if !exists {
			time.Sleep(10 * time.Second)
			response, err := client.Comment.Delete(ctx, "t1_"+comment.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"commentID": comment.ID,
					"error":     err,
				}).Error("Failed to delete comment")
				return
			}
			log.WithFields(log.Fields{
				"commentID": comment.ID,
				"response":  response.Response.Status,
				"rate":      response.Rate,
			}).Info("Deleted comment successfully")

			err = markCommentAsDeleted(db, comment.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"commentID": comment.ID,
					"error":     err,
				}).Error("Failed to mark comment as deleted")
				return
			}
		} else {
			log.WithField("commentID", comment.ID).Info("Comment already deleted")
		}
	}
}

func main() {
	cfg, err := config.ReadConfig("config/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Configuration loaded successfully")

	badgerDB, err := badger.NewBadgerDB(cfg.DB.DBFile)
	if err != nil {
		log.Fatalf("Failed to open BadgerDB: %s", err)
	}
	defer func() {
		if err := badgerDB.Close(); err != nil {
			log.Fatalf("Failed to close BadgerDB: %s", err)
		}
	}()

	credentials := reddit.Credentials{
		ID:       cfg.Reddit.ClientID,
		Secret:   cfg.Reddit.Secret,
		Username: cfg.Reddit.Username,
		Password: cfg.Reddit.Password,
	}
	client, err := reddit.NewClient(credentials)
	if err != nil {
		log.Fatalf("Failed to create Reddit client: %s", err)
	}

	ctx := context.Background()
	comments, _, err := client.User.Comments(ctx, &reddit.ListUserOverviewOptions{
		ListOptions: reddit.ListOptions{
			Limit: 1000,
		},
		Time: "all",
	})
	if err != nil {
		log.Fatalf("Failed to fetch comments: %s", err)
	}

	for _, comment := range comments {
		processComment(ctx, comment, client, badgerDB)
	}
}
