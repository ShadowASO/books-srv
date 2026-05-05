/*
---------------------------------------------------------------------------------------
File: pg_client.go
Autor: Aldenor
Data: 29-04-2026
---------------------------------------------------------------------------------------
//Conexão com o POSTGRESQL

	dbConfig := postgres.PGConfig{
		Host:     cfg.PgHost,
		Port:     cfg.PgPort,
		User:     cfg.PgUser,
		Password: cfg.PgPass,
		DBName:   cfg.PgDB,
		PoolSize: cfg.PGPoolSize,
	}
	db, err := postgres.NewPGConn(dbConfig)
	if err != nil {
		log.Fatalf("erro ao criar pool de conexões com o database: %v", err)
	}
	defer db.Close()
*/
package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PGPool struct {
	Pool *sql.DB
}

var PGPoolGlobal *PGPool

type PGConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	PoolSize int
}

func NewPGConn(cfg PGConfig) (*PGPool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	//Configurando o pool de conexões
	conn.SetMaxOpenConns(cfg.PoolSize)
	conn.SetMaxIdleConns(cfg.PoolSize)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	log.Printf("Conexão realizada com sucesso: %s", cfg.DBName)
	PGPoolGlobal = &PGPool{
		Pool: conn,
	}
	return PGPoolGlobal, nil
}

// Close cleanly shuts down the connection pool
func (db *PGPool) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
