package main

import (
	"fmt"
	"crypto/md5"
	"net/http"
	"github.com/mmcdole/gofeed"
)

func FetchSubscription(url string) (*Subscription, []*Item, error) {
	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
		return nil, nil, err
    }
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	subscription := &Subscription {
		id: fmt.Sprintf("%x", md5.Sum([]byte(feed.Link))),
		title: feed.Title,
		description: feed.Description,
		link: url,
	}

	var items []*Item
	for index := len(feed.Items) - 1; index >= 0; index-- {
		item := feed.Items[index]
		items = append(items, &Item {
			guid: item.GUID,
			title: item.Title,
			link: item.Link,
			date: item.PublishedParsed,
		})
	}

	return subscription, items, nil
}

func FetchItems(url string) ([]*Item, error) {
	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
		return nil, err
    }
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []*Item
	for index := len(feed.Items) - 1; index >= 0; index-- {
		item := feed.Items[index]
		items = append(items, &Item {
			guid: item.GUID,
			title: item.Title,
			link: item.Link,
			date: item.PublishedParsed,
		})
	}

	return items, nil
}