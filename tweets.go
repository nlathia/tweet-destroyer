package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func logTweet(state string, tweet twitter.Tweet) {
	tweetAge := ""
	created, err := tweet.CreatedAtTime()
	if err != nil {
		tweetAge = "failed to parse"
	} else {
		diff := time.Now().UTC().Sub(created).Hours() / 24
		tweetAge = fmt.Sprintf("%v days ago", diff)
	}

	log.Printf("%s: id=%s by %s (is_rt=%t, created=%s (%v), RT=%d, Fav=%d): %s",
		state,
		tweet.IDStr,
		tweet.User.ScreenName,
		(tweet.RetweetedStatus != nil),
		tweet.CreatedAt,
		tweetAge,
		tweet.RetweetCount,
		tweet.FavoriteCount,
		tweet.Text,
	)
}

func parseID(idStr string) (int64, error) {
	tweetID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return tweetID, nil
}

// getTwitterClient returns a new go-twitter client using secrets
// that have been stored in Google Cloud's Secret Manager
func getTwitterClient(ctx context.Context) (*twitter.Client, error) {
	secret, err := readSecret(ctx)
	if err != nil {
		return nil, err
	}

	config := oauth1.NewConfig(secret.ConsumerKey, secret.ConsumerSecret)
	token := oauth1.NewToken(secret.Token, secret.TokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	return twitter.NewClient(httpClient), nil
}

// getTweets returns a slice of tweets up to (and including) maxID.
func getTweets(client *twitter.Client, maxID int64) ([]twitter.Tweet, error) {
	log.Printf("retrieving tweets with maxID=%d", maxID)
	tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		// Count specifies the number of Tweets to try and retrieve, up to a maximum of
		// 200 per distinct request. The value of count is best thought of as a
		// limit to the number of Tweets to return because suspended or deleted content
		// is removed after the count has been applied. We include retweets in the count,
		// even if include_rts is not supplied.
		Count: 200,
		// MaxID returns results with an ID less than (that is, older than) *or equal*
		// to the specified ID.
		MaxID: maxID,
	})
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// getMinID returns the lowest (earliest) ID in the given slice of tweets
func getMinID(tweets []twitter.Tweet) (int64, error) {
	if len(tweets) == 0 {
		return 0, nil
	}

	minID, err := parseID(tweets[0].IDStr)
	if err != nil {
		return 0, err
	}
	for _, tweet := range tweets {
		tweetID, err := parseID(tweet.IDStr)
		if err != nil {
			return 0, err
		}
		if tweetID < minID {
			minID = tweet.ID
		}
	}
	log.Printf("new minID=%d", minID)
	return minID, nil
}

// filterTweets returns a slice of tweets that are candidates for deletion
func filterTweets(tweets []twitter.Tweet) ([]twitter.Tweet, error) {
	if len(tweets) == 0 {
		return []twitter.Tweet{}, nil
	}

	result := []twitter.Tweet{}
	for _, tweet := range tweets {
		delete, err := shouldDelete(tweet)
		if err != nil {
			return nil, err
		}
		if !delete {
			logTweet("keeping", tweet)
			continue
		}

		result = append(result, tweet)
	}

	log.Printf("found %d candidates to destroy", len(result))
	return result, nil
}

// deleteTweet destroys a given tweet and returns it if successful
func deleteTweet(client *twitter.Client, tweet twitter.Tweet) error {
	logTweet("destroying", tweet)
	tweetID, err := parseID(tweet.IDStr)
	if err != nil {
		return err
	}

	_, resp, err := client.Statuses.Destroy(tweetID, nil)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			log.Printf("already destroyed: id=%s (%d)", tweet.IDStr, tweet.ID)
			return nil
		}

		log.Printf("failed to call destroy (status=%d)", resp.StatusCode)
		return err
	}
	return nil
}

// deleteTweets destroys a slice of tweets and returns a count. Deletion will only
// happen if dryRun=false
func deleteTweets(client *twitter.Client, tweets []twitter.Tweet, dryRun bool) (int, error) {
	if len(tweets) == 0 {
		log.Printf("no tweets are eligible to be deleted")
		return 0, nil
	}

	numDeleted := 0
	for _, tweet := range tweets {
		if !dryRun {
			err := deleteTweet(client, tweet)
			if err != nil {
				log.Printf("error destroying id=%s, %v", tweet.IDStr, err.Error())
				return numDeleted, err
			}
		}
		numDeleted += 1
	}
	return numDeleted, nil
}
