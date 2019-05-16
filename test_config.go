package session

var TestConfig = &struct {
	RedisHost       string
	RedisPassword   string
	MySQLDatasource string
	FirestoreCreds  string
}{
	RedisHost:       "localhost:6379",
	RedisPassword:   "",
	MySQLDatasource: "[USERNAME]:[PASSWORD]@tcp([HOST]:3306)/[DATABASE]?parseTime=true",
	FirestoreCreds:  `{}`,
}
