### Development

This library is currently under development. Being ported out of an project specific library this is not quite ready for being imported.

### Current Support/Road Map

- [ ] File
- [x] Firestore
- [ ] Memcache
- [x] Memory
- [ ] MongoDB
- [x] MySQL
- [ ] PostgreSQL
- [x] Redis
- [ ] Encryption
- [ ] Cookie Sessions
- [x] Header Sessions

### Usage

Additional samples can be seen in `manager_test.go`

Basic session manager implementation

```
m := &Manager{
  Provider:    [PROVIDER],
  Lifetime:    session.DefaultSessionExpiration, // default 1200
}
```

If you would like to use the `HeaderLoad` or `CookieLoad` convenience methods specify the header key and type/prefix of the header key/cookie name.

```
m := &Manager{
  TokenName:   HeaderKey, // default Authorization
  TokenType:   HeaderPrefix, // default bearer
  Provider:    [PROVIDER],
  Lifetime:    session.DefaultSessionExpiration, // default 1200
}
```

## Providers

#### Memory
Memory provider is a simple key/value map held in memory for the duration of the execution of your application. 
```
p := session.NewMemoryProvider()
```

#### MySQL
MySQL provider stores the session data in a MySQL database. The MySQLProvider takes a db connection and the name of the table to store the session data.
```
db, err := sql.Open("mysql", "[CONNECTION_STRING]")
... err checks and defer close handling ...
p, err := session.NewMySQLProvider(db, session.DefaultMySQLTableName)
if err != nil {
  log.Fatal(err)
}
```

#### Redis
Redis provider stores the session data in a Redis store.  The RedisProvider takes the host, password, and db id.
```
p, _ := session.NewRedisProvider("[HOST]", "[PASSWORD]", [DATABASE])
```

#### Firestore
Firestore provider stores the session data in a Firestore database. It takes the current context, a firestore client, and the collection name
```
ctx := context.Background()

auth := option.WithCredentialsJSON([]byte("[FIRESTORE_CONFIG]"))
firebaseApp, err := firebase.NewApp(ctx, nil, auth)
... error checks and handling ...
fc, err := firebaseApp.Firestore(ctx)
... error checks and handling ...

p := session.NewFirestoreProvider(ctx, fc, "[COLLECTION_NAME]")
```