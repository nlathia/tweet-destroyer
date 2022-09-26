package main

import (
	"testing"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsOlderThan(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	testCases := []struct {
		created   time.Time
		threshold int
		isOlder   bool
	}{
		{
			created:   now.AddDate(0, 0, -1),
			threshold: 7,
			isOlder:   false,
		},
		{
			created:   now.AddDate(0, 0, -7),
			threshold: 7,
			isOlder:   true,
		},
		{
			created:   now.AddDate(0, 0, -30),
			threshold: 7,
			isOlder:   true,
		},
		{
			created:   now.AddDate(0, 0, -30),
			threshold: 30,
			isOlder:   true,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.isOlder, isOlderThan(testCase.created, testCase.threshold))
	}
}

func TestShouldDelete(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		tweet  twitter.Tweet
		delete bool
	}{
		{
			name: "not written by me",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "twitter",
				},
			},
			delete: false,
		},
		{
			name: "not written by me and is old",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().AddDate(0, 0, -30).Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "twitter",
				},
			},
			delete: true,
		},
		{
			name: "not older than a week and has no engagement",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().AddDate(0, 0, -1).Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "neal_lathia",
				},
			},
			delete: false,
		},
		{
			name: "older than a week and has no engagement",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().AddDate(0, 0, -7).Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "neal_lathia",
				},
			},
			delete: true,
		},
		{
			name: "older than a week and has engagement",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().AddDate(0, 0, -7).Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "neal_lathia",
				},
				RetweetCount: 1,
			},
			delete: false,
		},
		{
			name: "older than a month and has low engagement",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().AddDate(0, 0, -30).Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "neal_lathia",
				},
				RetweetCount: 1,
			},
			delete: true,
		},
		{
			name: "older than a month and has high engagement",
			tweet: twitter.Tweet{
				CreatedAt: time.Now().UTC().AddDate(0, 0, -30).Format("Mon Jan 02 15:04:05 -0700 2006"),
				User: &twitter.User{
					ScreenName: "neal_lathia",
				},
				RetweetCount:  5,
				FavoriteCount: 5,
			},
			delete: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			delete, err := shouldDelete(testCase.tweet)
			require.NoError(t, err)
			assert.Equal(t, testCase.delete, delete)
		})
	}
}
