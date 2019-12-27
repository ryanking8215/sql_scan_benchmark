package sqlscan

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	once     sync.Once
	dbName   = "file:sqlscan?mode=memory&cache=shared"
	dataSize = 100
)

func prepareDataOnce() {
	once.Do(func() {
		if err := prepareData(); err != nil {
			panic(err)
		}
	})
}

func prepareData() error {
	create := `
		CREATE TABLE user (
			id INTEGER PRIMARY KEY   AUTOINCREMENT,
			user_name VARCHAR(50) NULL,
			password VARCHAR(45) NOT NULL,
			email VARCHAR(50) NULL,
			mobile_phone VARCHAR(45) NULL,
			company VARCHAR(100) NULL,
			avatar_url VARCHAR(100) NULL DEFAULT '',
			role INT NULL DEFAULT 0,
			last_online_at INTEGER NOT NULL,
			create_at DATETIME NOT NULL,
			update_at DATETIME NOT NULL
		);
	`
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return err
	}
	// defer db.Close() // NOTE don't close, keep db exists if in memory mode !!!

	if _, err := db.Exec(create); err != nil {
		return err
	}

	for i := 0; i < dataSize; i++ {
		userName := fmt.Sprintf("user%d", i+1)
		password := fmt.Sprintf("password%d", i+1)
		email := fmt.Sprintf("user%d@sqlscan.com", i+1)
		mobilePhone := fmt.Sprintf("%d", 10000*(i+1))
		company := fmt.Sprintf("company%d", i+1)
		avatarURL := fmt.Sprintf("http://sqlscan.com/avatar/%d", i+1)
		role := i % 3
		lastOnlineAt := time.Now().Unix() + int64(i)
		createAt := time.Now().UTC()
		updateAt := time.Now().UTC()
		_, err := db.Exec(`INSERT INTO user (user_name, password, email, mobile_phone, company, avatar_url, role, last_online_at, create_at, update_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userName, password, email, mobilePhone, company, avatarURL, role, lastOnlineAt, createAt, updateAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func Test_PrepareData(t *testing.T) {
	err := prepareData()
	if err != nil {
		t.Fatal(err)
	}
}

type User struct {
	ID           int       `db:"id" gorm:"column:id" xorm:"'id'"`
	UserName     string    `db:"user_name" gorm:"column:user_name" xorm:"'user_name'"`
	Password     string    `db:"password" gorm:"column:password" xorm:"'password'"`
	Email        string    `db:"email" gorm:"column:email" xorm:"'email'"`
	MobilePhone  string    `db:"mobile_phone" gorm:"column:email" xorm:"'mobile_phone'"`
	Company      string    `db:"company" gorm:"column:company" xorm:"'company'"`
	AvatarURL    string    `db:"avatar_url" gorm:"column:avatar_url" xorm:"'avatar_url'"`
	Role         int       `db:"role" gorm:"column:role" xorm:"'role'"`
	LastOnlineAt int64     `db:"last_online_at" gorm:"column:last_online_at" xorm:"'last_online_at'"`
	CreateAt     time.Time `db:"create_at" gorm:"create_at" xorm:"'create_at'"`
	UpdateAt     time.Time `db:"update_at" gorm:"update_at" xorm:"'update_at'"`
}

func Test_SqlScan(t *testing.T) {
	prepareData()
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cnt := 0
	rows, err := db.Query("SELECT * FROM user")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		cnt++
	}
	if cnt != dataSize {
		t.Error("wrong cnt")
	}
}

func BenchmarkSqlScan(b *testing.B) {
	b.StopTimer()
	prepareDataOnce()
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rows, err := db.Query("SELECT * FROM user")
		if err != nil {
			panic(err)
		}
		var user User
		b.StartTimer()
		for rows.Next() {
			if err := rows.Scan(&user.ID, &user.UserName, &user.Password, &user.Email, &user.MobilePhone, &user.Company, &user.AvatarURL, &user.Role, &user.LastOnlineAt, &user.CreateAt, &user.UpdateAt); err != nil {
				panic(err)
			}
		}
		rows.Close()
	}
}

func BenchmarkSqlxScan(b *testing.B) {
	b.StopTimer()
	prepareDataOnce()
	db, err := sqlx.Connect("sqlite3", dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rows, err := db.Queryx("SELECT * FROM user")
		if err != nil {
			panic(err)
		}
		var user User
		b.StartTimer()
		for rows.Next() {
			if err := rows.StructScan(&user); err != nil {
				panic(err)
			}
		}
		rows.Close()
	}
}

func BenchmarkGormScan(b *testing.B) {
	b.StopTimer()
	prepareDataOnce()
	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rows, err := db.Raw("SELECT * FROM user").Rows()
		if err != nil {
			panic(err)
		}
		var user User
		b.StartTimer()
		for rows.Next() {
			if err := db.ScanRows(rows, &user); err != nil {
				panic(err)
			}
		}
		rows.Close()
	}
}

func BenchmarkXormScan(b *testing.B) {
	b.StopTimer()
	prepareDataOnce()
	db, err := xorm.NewEngine("sqlite3", dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var user User
		rows, err := db.Rows(&user)
		if err != nil {
			panic(err)
		}
		b.StartTimer()
		for rows.Next() {
			if err := rows.Scan(&user); err != nil {
				panic(err)
			}
		}
		rows.Close()
	}
}

func QueryToString(db *sql.DB, b *testing.B, sql string, args ...interface{}) (results [][]string, err error) {
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	cols, _ := rows.Columns()
	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}
	results = make([][]string, 0)

	i := 0

	b.StartTimer()

	for rows.Next() {
		if err := rows.Scan(scans...); err != nil {
			return nil, err
		}
		row := make([]string, 0)
		for _, v := range values { // append column value to one row
			row = append(row, string(v))
		}
		results = append(results, row) // append row to rows
		i++
	}
	rows.Close()

	b.StopTimer()

	return results, nil
}

func BenchmarkQueryToString(b *testing.B) {
	b.StopTimer()
	prepareDataOnce()
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		results, err := QueryToString(db, b, "SELECT * FROM user")
		if err != nil {
			panic(err)
		}
		var user User
		for _, res := range results {
			user.ID, _ = strconv.Atoi(res[0])
			user.UserName = res[1]
			user.Password = res[2]
			user.Email = res[3]
			user.MobilePhone = res[4]
			user.Company = res[5]
			user.AvatarURL = res[6]
			user.Role, _ = strconv.Atoi(res[7])
			user.LastOnlineAt, _ = strconv.ParseInt(res[8], 10, 64)
			user.CreateAt, _ = time.Parse(time.RFC3339, res[9])
			user.UpdateAt, _ = time.Parse(time.RFC3339, res[10])
		}
	}
}
