package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type searchResult struct {
	Items []item `json:"items"`
}

type item struct {
	User user `json:"user"`
}

type user struct {
	Login string `json:"login"`
}

func main() {
	author := flag.String("author", "", "the author to get failing PRs for")
	token := flag.String("token", "", "Github OAUTH token")
	flag.Parse()

	if *author == "" {
		fmt.Println("Please supply an author")
		return
	}

	if *token == "" {
		fmt.Println("Please supply an oauth token")
		return
	}

	client := &http.Client{}

	qs := "?q=type:pr+repo:nytm/np-well+state:open+status:failure+author:" + *author
	req, err := http.NewRequest("GET", "https://api.github.com/search/issues"+qs, nil)
	if err != nil {
		fmt.Println("error making new request", err)
		return
	}
	req.Header.Add("Authorization", "token "+*token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error doing request", err)
		return
	}

	var sr searchResult
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		fmt.Println("error decoding body", err)
		return
	}
	resp.Body.Close()

	if len(sr.Items) > 0 {
		fmt.Printf("User %s has a failing PR\n", *author)
	}
}
