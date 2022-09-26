package main

import (
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
func hasCountsBelow(tweet twitter.Tweet, threshold int) bool {
	numEngagements := tweet.FavoriteCount +
		tweet.RetweetCount +
		tweet.QuoteCount + //only available with the Premium and Enterprise tier products
		tweet.ReplyCount //only available with the Premium and Enterprise tier products
	return numEngagements < threshold
}

func shouldDelete(tweet twitter.Tweet) (bool, error) {
	created, err := tweet.CreatedAtTime()
	if err != nil {
		return false, err
	}

	// if I wrote the tweet
	if tweet.User.ScreenName == "neal_lathia" {
		if tweet.Favorited {
			// this Tweet has been liked by the authenticating user (me)
			// so we keep it
			return false, nil
		}

		switch {
		case isOlderThan(created, 30*6):
			// destroy anything older than 6 months with < 25 RTs/faves
			return hasCountsBelow(tweet, 25), nil
		case isOlderThan(created, 30): // one month old
			// destroy anything more than a month old with < 10 RTs/faves
			return hasCountsBelow(tweet, 10), nil
		case isOlderThan(created, 7):
			// destroy anything more than a week old with no engagement
			return hasCountsBelow(tweet, 1), nil
		default:
			// everything else is recent - keep it
			return false, nil
		}
	}

	return isOlderThan(created, 30), nil
}
