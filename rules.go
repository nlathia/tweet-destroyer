package main

import (
	"fmt"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

// isOlderThan is true if now - numDays is before the tweet's
// creation date
func isOlderThan(created time.Time, numDays int) bool {
	limit := time.Now().UTC().AddDate(0, 0, 0-numDays)
	return created.Before(limit)
}

// hasCountsBelow returns true if the sum of the engagement metrics
// for a tweet is below a given threshold
func hasCountsBelow(tweet twitter.Tweet, threshold int) string {
	numEngagements := tweet.FavoriteCount +
		tweet.RetweetCount +
		tweet.QuoteCount + //only available with the Premium and Enterprise tier products
		tweet.ReplyCount //only available with the Premium and Enterprise tier products
	if numEngagements < threshold {
		return ""
	}
	return fmt.Sprintf("engagement (%v >= %v)", numEngagements, threshold)
}

func reasonToKeep(tweet twitter.Tweet) (string, error) {
	created, err := tweet.CreatedAtTime()
	if err != nil {
		return "", err
	}

	// if I wrote the tweet
	// Retweets can be distinguished from typical Tweets by the existence of a
	// retweeted_status attribute. This attribute contains a representation of
	// the original Tweet that was retweeted.
	if tweet.RetweetedStatus == nil {
		if tweet.Favorited {
			// this Tweet has been liked by the authenticating user (me)
			// so we keep it
			return "self-fave", nil
		}

		thresholds := map[int]int{
			30 * 6: 25,
			30:     10,
			7:      1,
		}
		for age, threshold := range thresholds {
			if isOlderThan(created, age) {
				return hasCountsBelow(tweet, threshold), nil
			}
		}
		return "new-tweet", nil
	}

	// Delete all RTs after 30 days
	if isOlderThan(created, 30) {
		return "", nil
	}
	return "recent-rt", nil
}
