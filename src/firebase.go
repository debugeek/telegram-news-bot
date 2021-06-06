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

	iter := fb.firestore.Collection("accounts").Documents(fb.ctx)
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
	iter := fb.firestore.Collection("accounts").Where("id", "==", id).Documents(fb.ctx)

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

	_, err := fb.firestore.Collection("accounts").Doc(id).Set(fb.ctx, account)

	return err
}

// Subscription

func (fb Firebase) GetSubscriptions(account *Account) (map[string]*Subscription, error) {
	id := strconv.FormatInt(account.Id, 10)

	subscriptions := make(map[string]*Subscription)

	iter := fb.firestore.Collection("assets").Doc(id).Collection("subscriptions").Documents(fb.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var subscription Subscription
		err = doc.DataTo(&subscription)
		if err != nil {
			return nil, err
		}

		subscriptions[doc.Ref.ID] = &subscription
	}

	return subscriptions, nil
}

func (fb Firebase) AddSubscription(account *Account, subscription *Subscription) error {
	id := strconv.FormatInt(account.Id, 10)

	subscriptionRef := fb.firestore.Collection("assets").Doc(id).Collection("subscriptions").Doc(subscription.Id)
	statisticRef := fb.firestore.Collection("statistics").Doc("subscriptions").Collection("subscribe_count").Doc(subscription.Id)

	err := fb.firestore.RunTransaction(fb.ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var statistic SubscriptionStatistic

		doc, err := tx.Get(statisticRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				statistic = SubscriptionStatistic{
					Count:        0,
					Subscription: subscription,
				}
			} else {
				return err
			}
		} else {
			doc.DataTo(&statistic)
		}

		statistic.Count++

		err = tx.Set(statisticRef, statistic)
		if err != nil {
			return err
		}

		err = tx.Set(subscriptionRef, subscription)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (fb Firebase) DeleteSubscription(account *Account, subscription *Subscription) error {
	id := strconv.FormatInt(account.Id, 10)

	subscriptionRef := fb.firestore.Collection("assets").Doc(id).Collection("subscriptions").Doc(subscription.Id)
	statisticRef := fb.firestore.Collection("statistics").Doc("subscriptions").Collection("subscribe_count").Doc(subscription.Id)

	err := fb.firestore.RunTransaction(fb.ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var statistic SubscriptionStatistic

		doc, err := tx.Get(statisticRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				statistic = SubscriptionStatistic{
					Count:        0,
					Subscription: subscription,
				}
			} else {
				return err
			}
		} else {
			doc.DataTo(&statistic)
		}

		statistic.Count--

		if statistic.Count <= 0 {
			err = tx.Delete(statisticRef)
			if err != nil {
				return err
			}
		} else {
			err = tx.Set(statisticRef, statistic)
			if err != nil {
				return err
			}
		}

		err = tx.Delete(subscriptionRef)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func (fb Firebase) GetItemPushed(account *Account, itemId string) (bool, error) {
	id := strconv.FormatInt(account.Id, 10)

	dsnap, err := fb.firestore.Collection("assets").Doc(id).Collection("feeds").Doc(itemId).Get(fb.ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		} else {
			return false, err
		}
	} else {
		var results map[string]interface{}
		err = dsnap.DataTo(&results)
		if err != nil {
			return false, err
		}

		var pushed bool = (results["pushed"]).(bool)
		return pushed, nil
	}
}

func (fb Firebase) SetItemsPushed(account *Account, itemIds []string) error {
	id := strconv.FormatInt(account.Id, 10)

	batch := fb.firestore.Batch()

	for _, itemId := range itemIds {
		ref := fb.firestore.Collection("assets").Doc(id).Collection("feeds").Doc(itemId)
		batch.Set(ref, map[string]interface{}{
			"pushed": true,
		}, firestore.MergeAll)
	}

	_, err := batch.Commit(fb.ctx)

	return err
}

func (fb Firebase) GetTopSubscriptions(num int) ([]*SubscriptionStatistic, error) {
	statistics := make([]*SubscriptionStatistic, 0)

	err := fb.firestore.RunTransaction(fb.ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		querier := fb.firestore.Collection("statistics").Doc("subscriptions").Collection("subscribe_count").OrderBy("count", firestore.Desc).Limit(num)
		iter := tx.Documents(querier)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			var statistic SubscriptionStatistic
			err = doc.DataTo(&statistic)
			if err != nil {
				return err
			}

			statistics = append(statistics, &statistic)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return statistics, nil
}
