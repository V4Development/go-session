package session

var TestConfig = &struct{
	RedisPassword string
	MySQLDatasource string
	FirestoreCreds string
}{
	RedisPassword: "",
	MySQLDatasource: "[USERNAME]:[PASSWORD]@tcp([HOST]:3306)/[DATABASE]?parseTime=true",
	FirestoreCreds: `{}`,
}