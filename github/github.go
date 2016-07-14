package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Search struct {
	client *http.Client
	author string
	token  string
	repo   string
}

func NewSearch(author, token, repo string) *Search {
	return &Search{
		client: &http.Client{},
		author: author,
		token:  token,
		repo:   repo,
	}
}

type searchResult struct {
	Items []item `json:"items"`
}

type item struct {
	URL  string `json:"url"`
	User user   `json:"user"`
}

type user struct {
	Login string `json:"login"`
}

func (s *Search) Failing() (bool, error) {
	qs := "?q=type:pr+repo:" + s.repo + "+state:open+status:failure+author:" + s.author
	req, err := http.NewRequest("GET", "https://api.github.com/search/issues"+qs, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", "token "+s.token)
	resp, err := s.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("Bad response code %d", resp.StatusCode)
	}

	var sr searchResult
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return false, err
	}

	// TODO add logic to handle the edge case of a PR failing and
	// showing up under the status:failing criteria, and then a
	// new commit is pushed up which forces the tests to run again
	// and sets the status to pending, which makes this check look
	// like it's passing but really it might not be.
	if len(sr.Items) > 0 {
		// TODO push all items into a slice with PR #
		return true, nil
	}

	// TODO before marking all as passed, check any PRs in the
	// stored slice to make sure they're not pending. If they are
	// pending, don't mark as passed yet. If they're closed, remove
	// them from the slice.

	return false, nil
}
