// Package testdb abre um Postgres de teste isolado por schema.
//
// Os testes que o usam sao pulados sem TEST_DATABASE_URL, entao o CI segue
// sem banco. Localmente:
//
//	TEST_DATABASE_URL="postgres://postgres:senha@127.0.0.1:5544/todeolho?sslmode=disable" go test ./...
package testdb

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Abrir cria um schema novo, aplica o AutoMigrate dos modelos e o remove no fim do teste.
func Abrir(t *testing.T, modelos ...any) *gorm.DB {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL nao definido")
	}

	schema := fmt.Sprintf("t_%s_%d", strings.ToLower(strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())), time.Now().UnixNano()%1_000_000)
	admin, err := gorm.Open(postgres.Open(base), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()

	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		if sqlDB, err := admin.DB(); err == nil {
			sqlDB.Close()
		}
	})

	if err := db.AutoMigrate(modelos...); err != nil {
		t.Fatal(err)
	}
	return db
}
