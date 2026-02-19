package operations

import (
	"context"
	"eavesdropper/dtos/resources"
	"eavesdropper/services/data/firestore/collections"
	"errors"

	db "eavesdropper/services/data/firestore/client"

	"google.golang.org/api/iterator"
)

var dbContext = db.FirestoreInstance.Ctx

//////// WRITES ////////

func CreateUser(user *resources.User) error {
	_, err := collections.Users.Doc(user.ID).Create(context.Background(), user)
	return err
}

func UpdateUser(user *resources.User) error {
	_, err := collections.Users.Doc(user.ID).Set(context.Background(), user)
	return err
}

//////// READS ////////

func GetUser(id string) (*resources.User, error) {

	snap, err := collections.Users.Doc(id).Get(dbContext)
	if err != nil {
		return nil, err
	}

	user := new(resources.User)
	err = snap.DataTo(user)

	return user, err
}

func HandleIsAvailable(ctx context.Context, handle string) (bool, error) {
	iter := collections.Users.
		Where("Handle", "==", handle).
		Limit(1).
		Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return false, err
		}
		return !doc.Exists(), nil
	}
	return true, nil
}

func GetUserWithStripe(stripeID string) (*resources.User, error) {

	ctx := context.Background()

	snaps, err := collections.Users.Where("StripeCustomerID", "==", stripeID).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, errors.New("No user found.")
	}
	if len(snaps) != 1 {
		return nil, errors.New("Multiple users found with the same stripe id")
	}

	user := new(resources.User)
	err = snaps[0].DataTo(user)

	return user, err
}
