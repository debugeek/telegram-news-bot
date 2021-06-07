package main

import (
	"context"
	"encoding/base64"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
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
