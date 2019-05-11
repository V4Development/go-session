package provider

import (
	"cloud.google.com/go/firestore"
	"context"
	"google.golang.org/api/iterator"
	"log"
	"time"
)

// Firebase Session Provider

const DefaultFirestoreCollection = "session"

type FirestoreProvider struct {
	Context context.Context
	Client *firestore.Client
	Collection *firestore.CollectionRef
	CollectionName string
}

func NewFirestoreProvider(ctx context.Context, client *firestore.Client, collectionName string) *FirestoreProvider {
	collection := client.Collection(collectionName)

	return &FirestoreProvider{
		ctx,
		client,
		collection,
		collectionName,
	}
}

func (p *FirestoreProvider) Read(sid string) (*Session, error) {
	doc, err := p.Collection.Doc(sid).Get(p.Context)
	if err != nil {
		return nil, err
	}

	var sess Session
	if err := doc.DataTo(&sess); err != nil {
		return nil, err
	}

	return &sess, nil
}

func (p *FirestoreProvider) Save(session *Session) error {
	_, err := p.Collection.Doc(session.UUID).Set(p.Context, session)
	if err != nil {
		log.Print(err)
	}
	return err
}

func (p *FirestoreProvider) Destroy(sid string) error {
	doc, err := p.Collection.Doc(sid).Get(p.Context)
	if err != nil {
		return err
	}

	if _, err := doc.Ref.Delete(p.Context); err != nil {
		return err
	}

	return nil
}

func (p *FirestoreProvider) GarbageCollect() {
	q := p.Collection.Where("Expire", "<", time.Now())
	iter := q.Documents(p.Context)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Print(err)
		}

		if _, err := doc.Ref.Delete(p.Context); err != nil {
			log.Print(err)
		}
	}

}
