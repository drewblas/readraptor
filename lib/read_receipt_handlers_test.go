package readraptor

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/asm-products/readraptor/fake"

	"github.com/codegangsta/martini"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

func Test_Tracking(t *testing.T) {
	// initialize the DbMap
	dbmap := initTestDb(t)
	defer dbmap.Db.Close()

	// delete any existing rows
	err := dbmap.TruncateTables()
	checkErr(t, err, "TruncateTables failed")

	account := NewAccount("weasley@example.com")
	err = dbmap.Insert(account)
	checkErr(t, err, "Insert failed")

	params := martini.Params{
		"public_key": account.PublicKey,
		"article_id": "article_1",
		"user_id":    "user_1",
		"signature":  signature(account.PrivateKey, account.PublicKey, "article_1", "user_1"),
	}

	req := &http.Request{}
	req.URL, _ = url.Parse("/t")
	rw := fake.New(t)
	GetTrackReadReceipts("..")(params, rw, req)

	gif, _ := ioutil.ReadFile("../public/tracking.gif")
	rw.Assert(200, gif)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func initTestDb(t *testing.T) *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	InitDb("postgres://localhost/rr_test?sslmode=disable")

	err := dbmap.CreateTablesIfNotExists()
	checkErr(t, err, "Create tables failed")

	return dbmap
}

func checkErr(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s – %s", err, message)
	}
}
