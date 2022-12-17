package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type deleteRequest struct {
	DryRun  bool `json:"dry_run"`        // if true, does not delete tweets
	MaxIter int  `json:"max_iterations"` // if set, caps the # of batches of tweets that are collected
}

type deleteResponse struct {
	DryRun       bool   `json:"dry_run"`       // if true, no tweets actually deleted
	NumDeleted   int    `json:"num_deleted"`   // # tweets that were (would have been) deleted
	NumCollected int    `json:"num_collected"` // # tweets that were retrieved
	Error        string `json:"error"`         // first error encountered with the twitter API
}

// handleDeleteTweets collects tweets and deletes the ones that match against
// a set of rules
func handleDeleteTweets(w http.ResponseWriter, r *http.Request) {
	// Read and validate the request
	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a client to query the twitter API
	twitterClient, err := getTwitterClient(r.Context(), nil)
	if err != nil {
		log.Printf("%v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// minID is used to query the Twitter API for results with an
	// ID less than (that is, older than) or equal to the specified ID
	minID := int64(0)
	iter := 0

	// We keep running counters for the response
	rsp := deleteResponse{
		DryRun: req.DryRun,
	}

	// Iterate over tweets and find candidates to delete
	for {

		// If max_iterations is given, check whether we should stop
		if req.MaxIter != 0 { // Only if MaxIter has been given
			if req.MaxIter == iter {
				break
			}
			iter += 1
		}

		// Retrieve a batch of tweets up to and including minID
		tweets, err := getTweets(twitterClient, minID)
		if err != nil {
			log.Printf("%v", err.Error())
			rsp.Error = err.Error()

			// Note: the run could have successfully deleted tweets
			// on a previous iteration of the loop
			// so we "successfully fail"
			break
		}

		if len(tweets) == 0 {
			log.Printf("no tweets retrieved (minID=%d)", minID)
			break
		}
		rsp.NumCollected += len(tweets)

		// Find and set the new minID
		newMinID, err := getMinID(tweets)
		if err != nil {
			log.Printf("failed to parse id: %s", err.Error())
			rsp.Error = err.Error()
			break
		}
		if minID == newMinID {
			log.Printf("no new min ID retrieved (%d)", minID)
			break
		}
		minID = newMinID

		// Filter and destroy selected tweets
		tweetsToDelete, err := filterTweets(tweets)
		if err != nil {
			rsp.Error = err.Error()
			break
		}

		numDeleted, err := deleteTweets(twitterClient, tweetsToDelete, req.DryRun)
		rsp.NumDeleted += numDeleted // incrementing before handling the err to account for partial success
		if err != nil {
			rsp.Error = err.Error()
			break
		}
	}

	// Format and write the result
	rspJSON, err := json.Marshal(rsp)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(rspJSON)
}
