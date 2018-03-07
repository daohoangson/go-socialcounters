// +build !appengine

package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/bmizerany/mc"
	"github.com/rubenv/sql-migrate"
)

const MYSQL_TABLE_NAME_HISTORY = "sc_history"
const MYSQL_COLUMN_NAME_SERVICE = "service"
const MYSQL_COLUMN_NAME_URL = "url"
const MYSQL_COLUMN_NAME_COUNT = "count"
const MYSQL_COLUMN_NAME_TIME = "time"

type Other struct {
}

func OtherNew(r *http.Request) Utils {
	utils := new(Other)

	return utils
}

var httpClient = &http.Client{
	Timeout: 1 * time.Second,
}

func (u Other) HttpGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (u Other) ConfigSet(key string, value string) error {
	return errors.New("Not implemented")
}

func (u Other) ConfigGet(key string) string {
	return os.Getenv(key)
}

func (u Other) Delay(handlerName string, args ...interface{}) error {
	handler, ok := DelayHandlers[handlerName]
	if !ok {
		return errors.New(fmt.Sprintf("Handler %s could not be found", handlerName))
	}

	// TODO: use thread pool / channel
	go func() {
		Verbosef(u, "Other.Delay: executing %s(%v)", handlerName, &args)
		handler(u, args...)
	}()

	return nil
}

func (u Other) MemorySet(items *[]MemoryItem) error {
	if items == nil || len(*items) < 1 {
		return nil
	}

	conn := getMcConn(u)
	if conn == nil {
		return errors.New("No memcache connection")
	}

	for _, item := range *items {
		if err := conn.Set(item.Key, item.Value, 0, 0, int(item.Ttl)); err != nil {
			u.Errorf("conn.Set(%s) error %v", item.Key, err)
		}
	}

	return nil
}

func (u Other) MemoryGet(items *[]MemoryItem) error {
	if items == nil || len(*items) < 1 {
		return nil
	}

	conn := getMcConn(u)
	if conn == nil {
		return errors.New("No memcache connection")
	}

	for index, item := range *items {
		if value, _, _, err := conn.Get(item.Key); err != nil {
			Verbosef(u, "conn.Get(%s) error %v", item.Key, err)
		} else {
			(*items)[index].Value = value
		}
	}

	return nil
}

func (u Other) HistorySave(records *[]HistoryRecord) error {
	if records == nil || len(*records) < 1 {
		return nil
	}

	conn := getDbConn(u)
	if conn == nil {
		return errors.New("No db connection")
	}

	sqlStr := "INSERT INTO " + MYSQL_TABLE_NAME_HISTORY +
		" (" + MYSQL_COLUMN_NAME_SERVICE +
		", " + MYSQL_COLUMN_NAME_URL +
		", " + MYSQL_COLUMN_NAME_COUNT +
		", " + MYSQL_COLUMN_NAME_TIME +
		") VALUES "
	vals := []interface{}{}
	for _, record := range *records {
		sqlStr += "(?, ?, ?, ?),"
		vals = append(vals, record.Service, record.Url, record.Count, record.Time)
	}
	sqlStr = sqlStr[:len(sqlStr)-1]
	Verbosef(u, "Other.HistorySave sqlStr = %s, vals = %v", sqlStr, vals)

	stmt, err := conn.Prepare(sqlStr)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}

	return nil
}

func (u Other) HistoryLoad(url string) ([]HistoryRecord, error) {
	records := []HistoryRecord{}

	conn := getDbConn(u)
	if conn == nil {
		return records, errors.New("No db connection")
	}

	sqlStr := "SELECT " + MYSQL_COLUMN_NAME_SERVICE +
		", " + MYSQL_COLUMN_NAME_URL +
		", " + MYSQL_COLUMN_NAME_COUNT +
		", " + MYSQL_COLUMN_NAME_TIME +
		" FROM " + MYSQL_TABLE_NAME_HISTORY +
		" WHERE " + MYSQL_COLUMN_NAME_URL + " = ?"
	stmt, err := conn.Prepare(sqlStr)
	if err != nil {
		return records, err
	}

	rows, err := stmt.Query(url)
	if err != nil {
		return records, err
	}
	defer rows.Close()

	for rows.Next() {
		var r HistoryRecord
		if err := rows.Scan(&r.Service, &r.Url, &r.Count, &r.Time); err != nil {
			return records, err
		}

		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return records, err
	}

	return records, nil
}

func (u Other) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (u Other) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (u Other) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

var mcConn *mc.Conn
var mcPrepared = false

func getMcConn(u Other) *mc.Conn {
	if !mcPrepared {
		if addr := os.Getenv("MEMCACHIER_SERVERS"); addr != "" {
			if m, err := mc.Dial("tcp", addr); err == nil {
				username := os.Getenv("MEMCACHIER_USERNAME")
				password := os.Getenv("MEMCACHIER_PASSWORD")

				if username != "" && password != "" {
					// only try to authenticate if both username and password are set
					err = m.Auth(os.Getenv("MEMCACHIER_USERNAME"), os.Getenv("MEMCACHIER_PASSWORD"))
					if err == nil {
						u.Infof("Other.getMcConn: mc.Auth ok")
						mcConn = m
					} else {
						u.Errorf("Other.getMcConn: mc.Auth error %v", err)
					}
				} else {
					// most of the case, the server does not require authentication
					u.Infof("Other.getMcConn: mc.Dial ok")
					mcConn = m
				}
			} else {
				u.Errorf("Other.getMcConn: mc.Dial error %v", err)
			}
		}

		mcPrepared = true
	}

	return mcConn
}

var dbConn *sql.DB
var dbPrepared = false

func getDbConn(u Other) *sql.DB {
	if !dbPrepared {
		if url := os.Getenv("CLEARDB_DATABASE_URL"); url != "" {
			dsn := url
			typePrefix := "mysql://"
			if strings.Index(dsn, typePrefix) == 0 {
				dsn = dsn[len(typePrefix):]
			}
			u.Debugf("Other.getDbConn: dsn=%s", dsn)

			if conn, err := sql.Open("mysql", dsn); err == nil {
				if err := conn.Ping(); err == nil {
					dbConn = conn
					u.Infof("Other.getDbConn: conn.Ping ok")

					verifyDbSchema(u)
				} else {
					u.Errorf("Other.getDbConn: conn.Ping error %v", err)
				}
			} else {
				u.Errorf("Other.getDbConn: sql.Open(%s) error %v", dsn, err)
			}
		}

		dbPrepared = true
	}

	return dbConn
}

func verifyDbSchema(u Other) {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			&migrate.Migration{
				Id: "1",
				Up: []string{
					"CREATE TABLE " + MYSQL_TABLE_NAME_HISTORY +
						" (" +
						MYSQL_COLUMN_NAME_SERVICE + " VARCHAR(255) NOT NULL, " +
						MYSQL_COLUMN_NAME_URL + " VARCHAR(255) NOT NULL, " +
						MYSQL_COLUMN_NAME_COUNT + " INT(10) UNSIGNED NOT NULL, " +
						MYSQL_COLUMN_NAME_TIME + " DATETIME NOT NULL, " +
						"KEY key_url  (" + MYSQL_COLUMN_NAME_URL + "), " +
						"KEY key_time (" + MYSQL_COLUMN_NAME_TIME + ")" +
						")",
				},
				Down: []string{"DROP TABLE " + MYSQL_TABLE_NAME_HISTORY},
			},
		},
	}

	n, err := migrate.Exec(dbConn, "mysql", migrations, migrate.Up)
	if err != nil {
		u.Errorf("migrate.Exec error %v", err)
		return
	}

	Verbosef(u, "Other.verifyDbSchema: migrate.Exec ok n=%d", n)
}
