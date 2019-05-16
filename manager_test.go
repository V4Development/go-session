package session

import (
	"context"
	"database/sql"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/v4development/go-session/session"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestSession(t *testing.T) {

}

func TestMemoryProvider(t *testing.T) {
	m := &Manager{
		KeyName:  DefaultHeaderKey,
		KeyType:  DefaultHeaderType,
		Provider: session.NewMemoryProvider(),
		Lifetime: session.DefaultSessionExpiration,
	}

	runTest(m, 0)
}

func TestRedisProvider(t *testing.T) {
	p, _ := session.NewRedisProvider(TestConfig.RedisHost, TestConfig.RedisPassword, session.DefaultRedisDatabase)

	m := &Manager{
		KeyName:  DefaultHeaderKey,
		KeyType:  DefaultHeaderType,
		Provider: p,
		Lifetime: session.DefaultSessionExpiration,
	}

	runTest(m, 10)
}

func TestMySQLProvider(t *testing.T) {
	// database connection
	db, err := sql.Open("mysql", TestConfig.MySQLDatasource)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		fmt.Println("closing database...")
		if err := db.Close(); err != nil {
			fmt.Println("error closing db: " + err.Error())
		}
	}()

	p, err := session.NewMySQLProvider(db, session.DefaultMySQLTableName)
	if err != nil {
		t.Error(err)
	}

	m := &Manager{
		KeyName:  DefaultHeaderKey,
		KeyType:  DefaultHeaderType,
		Provider: p,
		Lifetime: session.DefaultSessionExpiration,
	}

	runTest(m, 10)
}

func TestFirestoreProvider(t *testing.T) {
	ctx := context.Background()

	// auth and connect with firebase
	auth := option.WithCredentialsJSON([]byte(TestConfig.FirestoreCreds))
	firebaseApp, err := firebase.NewApp(ctx, nil, auth)
	if err != nil {
		log.Fatal(err)
	}

	// setup firestore client
	fc, err := firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatal(err)
	}

	p := session.NewFirestoreProvider(ctx, fc, session.DefaultFirestoreCollection)

	m := &Manager{
		KeyName:  DefaultHeaderKey,
		KeyType:  DefaultHeaderType,
		Provider: p,
		Lifetime: session.DefaultSessionExpiration,
	}

	runTest(m, 10)
}

func runTest(manager *Manager, deleteDelay time.Duration) {
	session := manager.NewSession()
	sid := session.UUID

	fmt.Println("********* Save Empty *************")
	err := manager.Save(session)
	if err != nil {
		fmt.Print("err")
		fmt.Println(err)
	}
	fmt.Println("*************************")

	fmt.Println("************* Read ************")
	sess, err := manager.Provider.Read(sid)
	if err != nil {
		fmt.Println("err")
		fmt.Println(err)
	}

	j, err := json.Marshal(sess)
	fmt.Println(string(j))
	fmt.Println("*************************")

	fmt.Println("************ HTTP Request Load *************")
	request := &http.Request{
		Header: make(http.Header),
	}
	request.Header.Set(DefaultHeaderKey, DefaultHeaderType+" "+sid)

	sess, err = manager.HeaderLoad(request)
	if err != nil {
		fmt.Println(err)
	}

	j, err = json.Marshal(sess)
	fmt.Println(string(j))
	fmt.Println("*************************")

	fmt.Println("********* Save and Load Update *************")
	sess.Data["value-string"] = "Test Data String"
	sess.Data["value-int"] = 100
	sess.Data["value-float"] = 100.001

	err = manager.Save(sess)
	if err != nil {
		fmt.Print("err")
		fmt.Println(err)
	}

	fmt.Println("************ Load Updated *************")
	request = &http.Request{
		Header: make(http.Header),
	}
	request.Header.Set(DefaultHeaderKey, DefaultHeaderType+" "+sid)

	sess, err = manager.HeaderLoad(request)
	if err != nil {
		fmt.Println(err)
	}

	j, err = json.Marshal(sess)
	fmt.Println(string(j))
	fmt.Println("*************************")

	fmt.Println("********* Destroy *************")
	time.Sleep(deleteDelay * time.Second)
	manager.Destroy(sess)

	fmt.Println("************ Load Destroyed *************")
	request = &http.Request{
		Header: make(http.Header),
	}
	request.Header.Set(DefaultHeaderKey, DefaultHeaderType+" "+sess.UUID)

	sess, err = manager.HeaderLoad(request)
	if err != nil {
		fmt.Println(err)
		fmt.Println("  -- expected")
	} else {
		j, err = json.Marshal(sess)
		fmt.Println(string(j))
	}

	fmt.Println("*************************")

	fmt.Println("********* Garbage Collect *************")
	manager.GarbageCollect()
	fmt.Println("*************************")
}
