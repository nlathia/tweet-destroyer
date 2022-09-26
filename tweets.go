package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

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
	tweets, resp, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
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
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response: %d", resp.StatusCode)
	}
	return tweets, nil
}

// getMinID returns the lowest (earliest) ID in the given slice of tweets
func getMinID(tweets []twitter.Tweet) int64 {
	if len(tweets) == 0 {
		return 0
	}

	minID := tweets[0].ID
	for _, tweet := range tweets {
		if tweet.ID < minID {
			minID = tweet.ID
		}
	}
	return minID
}

// filterTweets returns a slice of tweets that are candidates for deletion
func filterTweets(tweets []twitter.Tweet) ([]*twitter.Tweet, error) {
	if len(tweets) == 0 {
		return []*twitter.Tweet{}, nil
	}

	result := []*twitter.Tweet{}
	for _, tweet := range tweets {
		delete, err := shouldDelete(tweet)
		if err != nil {
			return nil, err
		}
		if delete {
			result = append(result, &tweet)
		}
	}
	return result, nil
}

// deleteTweet destroys a given tweet and returns it if successful
func deleteTweet(client *twitter.Client, tweet *twitter.Tweet) (*twitter.Tweet, error) {
	tweet, resp, err := client.Statuses.Destroy(tweet.ID, &twitter.StatusDestroyParams{})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response: %d", resp.StatusCode)
	}
	return tweet, nil
}

// deleteTweets destroys a slice of tweets and returns a count. Deletion will only
// happen if dryRun=false
func deleteTweets(client *twitter.Client, tweets []*twitter.Tweet, dryRun bool) (int, error) {
	numDeleted := 0
	var err error
	for _, tweet := range tweets {
		if !dryRun {
			_, err = deleteTweet(client, tweet)
			if err != nil {
				log.Printf("%v", err.Error())
				break
			}

			// TODO? - store the deleted tweet
		}
		numDeleted += 1
	}
	return numDeleted, err
}
