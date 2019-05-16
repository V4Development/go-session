package session

import (
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"log"
	"sync"
)

const DefaultMySQLTableName = "session"

type MySQLProvider struct {
	*sql.DB
	Table string
}

func NewMySQLProvider(db *sql.DB, table string) (*MySQLProvider, error) {
	p := &MySQLProvider{
		db,
		table,
	}

	if err := p.MySQLSetupCheck(); err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return p, nil
}

func (p *MySQLProvider) Read(sid string) (*Session, error) {
	var sess Session
	var d []byte
	q := "SELECT * FROM " + p.Table + " WHERE uuid=?"
	row := p.QueryRow(q, sid)
	if err := row.Scan(&sess.UUID, &d, &sess.Expire); err != nil {
		return nil, err
	}

	err := json.Unmarshal(d, &sess.Data)
	if err != nil {
		return nil, err
	}

	sess.Lock = sync.Mutex{}

	return &sess, nil
}

func (p *MySQLProvider) Save(session *Session) error {
	q := "INSERT INTO " + p.Table + " SET uuid=?, data=?, expire=? " +
		"ON DUPLICATE KEY UPDATE data=?, expire=?"
	stmt, err := p.Prepare(q)
	if err != nil {
		return err
	}

	data, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(session.UUID, data, session.Expire, data, session.Expire)
	if err != nil {
		return err
	}

	return nil
}

func (p *MySQLProvider) Destroy(sid string) error {
	q := "DELETE FROM " + p.Table + " WHERE uuid=?"
	stmt, err := p.Prepare(q)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(sid)
	if err != nil {
		return err
	}

	return nil
}

func (p *MySQLProvider) GarbageCollect() {
	q := "DELETE FROM " + p.Table + " WHERE expire < CURRENT_TIMESTAMP"
	stmt, err := p.Prepare(q)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//exp := time.Now()
	if _, err = stmt.Exec(); err != nil {
		log.Println(err.Error())
	}
}

func (p *MySQLProvider) MySQLSetupCheck() error {
	q := "DESCRIBE " + p.Table
	if _, err := p.Exec(q); err != nil {
		if e, ok := err.(*mysql.MySQLError); ok && e.Number == 1146 {
			// create the table
			q = "CREATE TABLE IF NOT EXISTS " + p.Table + " (" +
				"uuid varchar(36) not null," +
				"data blob null," +
				"expire datetime default CURRENT_TIMESTAMP not null," +
				"constraint " + p.Table + "_pk " +
				"primary key (uuid))"

			_, err := p.Exec(q)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
