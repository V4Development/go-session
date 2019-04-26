package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"testing"
)

func TestSession(t *testing.T) {

}

func TestRedisProvider(t *testing.T) {
	manager := Manager{
		SessionType: TypeHeader, //not really needed. all of them are headers
		TokenName:   HeaderKey,
		TokenType:   HeaderPrefix,
		Provider: &RedisProvider{
			Config: &RedisConfig{
				Server:   "localhost:6379",
				Password: "",
				Database: DefaultRedisDatabase,
			},
		},
		Lifetime: DefaultSessionExpire,
	}

	sid := manager.UUID()
	fmt.Println(sid)

	session := manager.Provider.Init(sid)
	session.Data["value-string"] = "Test Data String"
	session.Data["value-int"] = 100
	session.Data["value-float"] = 100.001

	err := manager.Save(session)
	if err != nil {
		fmt.Println(err)
	}

	sess, err := manager.Provider.Read(sid)
	if err != nil {
		fmt.Println(err)
	}

	j, err := json.Marshal(sess)
	fmt.Println(string(j))

	request := &http.Request{
		Header: make(http.Header),
	}
	request.Header.Set(HeaderKey, HeaderPrefix+" "+sid)

	sess, err = manager.RequestLoad(request)
	if err != nil {
		fmt.Println(err)
	}

	j, err = json.Marshal(sess)
	fmt.Println(string(j))
}

func TestMySQLProvider(t *testing.T) {
	// database connection
	ds := "[USERNAME]:[PASSWORD]@tcp([HOST]:3306)/[DATABASE]?parseTime=true"
	db, err := sql.Open("mysql", ds)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		fmt.Println("closing database...")
		if err := db.Close(); err != nil {
			fmt.Println("error closing db: " + err.Error())
		}
	}()

	provider := &MySQLProvider{
		DB:    db,
		Table: DefaultMySQLTableName,
	}
	err = provider.MySQLInit()
	if err != nil {
		t.Error(err)
	}

	manager := Manager{
		SessionType: TypeHeader, //not really needed. all of them are headers
		TokenName:   HeaderKey,
		TokenType:   HeaderPrefix,
		Provider:    provider,
		Lifetime:    DefaultSessionExpire,
	}

	sid := manager.UUID()
	fmt.Println(sid)

	session := manager.Provider.Init(sid)

	fmt.Println("********* Save Empty *************")
	err = manager.Save(session)
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

	fmt.Println("************ Load *************")
	request := &http.Request{
		Header: make(http.Header),
	}
	request.Header.Set(HeaderKey, HeaderPrefix+" "+sid)

	sess, err = manager.RequestLoad(request)
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
	request.Header.Set(HeaderKey, HeaderPrefix+" "+sid)

	sess, err = manager.RequestLoad(request)
	if err != nil {
		fmt.Println(err)
	}

	j, err = json.Marshal(sess)
	fmt.Println(string(j))
	fmt.Println("*************************")

	fmt.Println("********* Destroy *************")
	manager.Destroy(sess)

	fmt.Println("************ Load Destroyed *************")
	request = &http.Request{
		Header: make(http.Header),
	}
	request.Header.Set(HeaderKey, HeaderPrefix+" "+sess.UUID)

	sess, err = manager.RequestLoad(request)
	if err != nil {
		fmt.Println(err)
	} else {
		j, err = json.Marshal(sess)
		fmt.Println(string(j))
	}

	fmt.Println("*************************")
}
