package challenge

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	challengestore "github.com/abhimanyu95/scep/v2/challenge/bolt"
	scepserver "github.com/abhimanyu95/scep/v2/server"

	"github.com/boltdb/bolt"
	"github.com/smallstep/scep"
)

func TestDynamicChallenge(t *testing.T) {
	db, err := openTempBolt("scep-challenge")
	if err != nil {
		t.Fatal(err)
	}

	depot, err := challengestore.NewBoltDepot(db)
	if err != nil {
		t.Fatal(err)
	}

	// use the exported interface
	store := Store(depot)

	// get first challenge
	challengePassword, err := store.SCEPChallenge()
	if err != nil {
		t.Fatal(err)
	}

	if challengePassword == "" {
		t.Error("empty challenge returned")
	}

	// test store API
	valid, err := store.HasChallenge(challengePassword)
	if err != nil {
		t.Fatal(err)
	}
	if valid != true {
		t.Error("challenge just acquired is not valid")
	}
	valid, err = store.HasChallenge(challengePassword)
	if err != nil {
		t.Fatal(err)
	}
	if valid != false {
		t.Error("challenge should not be valid twice")
	}

	// get another challenge
	challengePassword, err = store.SCEPChallenge()
	if err != nil {
		t.Fatal(err)
	}

	if challengePassword == "" {
		t.Error("empty challenge returned")
	}

	// test CSRSigner middleware
	signer := Middleware(depot, scepserver.NopCSRSigner())

	csrReq := &scep.CSRReqMessage{
		ChallengePassword: challengePassword,
	}

	ctx := context.Background()

	_, err = signer.SignCSRContext(ctx, csrReq)
	if err != nil {
		t.Error(err)
	}

	_, err = signer.SignCSRContext(ctx, csrReq)
	if err == nil {
		t.Error("challenge should not be valid twice")
	}

}

func openTempBolt(prefix string) (*bolt.DB, error) {
	f, err := ioutil.TempFile("", prefix+"-")
	if err != nil {
		return nil, err
	}
	f.Close()
	err = os.Remove(f.Name())
	if err != nil {
		return nil, err
	}

	return bolt.Open(f.Name(), 0644, nil)
}
