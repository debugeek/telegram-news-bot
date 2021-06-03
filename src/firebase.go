package main

import (
	"context"
	"encoding/base64"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Firebase struct {
	app       *firebase.App
	firestore *firestore.Client
	ctx       context.Context
}

func SharedFirebase() Firebase {
	firebaseOnce.Do(func() {
		fb = Firebase{}

		fb.ctx = context.Background()

		var err error
		var opt option.ClientOption

		if FileExists("./credentials/serviceAccountKey.json") {
			opt = option.WithCredentialsFile("./credentials/serviceAccountKey.json")
		} else if cfg, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_SERVICE_ACCOUNT_BASE64")); err == nil {
			opt = option.WithCredentialsJSON(cfg)
		} else {
			panic("can't find google account credential")
		}

		fb.app, err = firebase.NewApp(fb.ctx, nil, opt)
		if err != nil {
			panic(err)
		}

		fb.firestore, err = fb.app.Firestore(fb.ctx)
		if err != nil {
			panic(err)
		}
	})
	return fb
}

// Account

func (fb Firebase) GetAccounts() ([]*Account, error) {
	accounts := make([]*Account, 0)

	iter := fb.firestore.Collection("account").Documents(fb.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var account Account
		doc.DataTo(&account)

		accounts = append(accounts, &account)
	}

	return accounts, nil
}

func (fb Firebase) GetAccount(id int64) (*Account, error) {
	iter := fb.firestore.Collection("account").Where("id", "==", id).Documents(fb.ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var account *Account
	err = doc.DataTo(&account)

	return account, err
}

func (fb Firebase) SaveAccount(account *Account) error {
	id := strconv.FormatInt(account.Id, 10)

	_, err := fb.firestore.Collection("account").Doc(id).Set(fb.ctx, account)

	return err
}

// Subscription

func (fb Firebase) GetSubscription(account *Account) (*Subscription, error) {
	id := strconv.FormatInt(account.Id, 10)

	dsnap, err := fb.firestore.Collection("subscription").Doc(id).Get(fb.ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var subscription Subscription
	err = dsnap.DataTo(&subscription)
	return &subscription, err
}

func (fb Firebase) SaveSubscription(account *Account, subscription *Subscription) error {
	id := strconv.FormatInt(account.Id, 10)

	_, err := fb.firestore.Collection("subscription").Doc(id).Set(fb.ctx, subscription)

	return err
}

func (fb Firebase) PostRecords(account *Account) (map[string]bool, error) {
	id := strconv.FormatInt(account.Id, 10)

	var records map[string]bool

	dsnap, err := fb.firestore.Collection("post_record").Doc(id).Get(fb.ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			records = make(map[string]bool)
		} else {
			return nil, err
		}
	} else {
		err = dsnap.DataTo(&records)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}

func (fb Firebase) MarkItemsPosted(account *Account, items []*Item) error {
	records, err := fb.PostRecords(account)
	if err != nil {
		return err
	}

	for _, item := range items {
		records[item.guid] = true
	}

	id := strconv.FormatInt(account.Id, 10)

	_, err = fb.firestore.Collection("post_record").Doc(id).Set(fb.ctx, records)

	return err
}
