package models

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/diadata-org/diadata/pkg/dia"
	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

// RelDatastore is a (persistent) relational database with an additional redis caching layer
type RelDatastore interface {

	// NFT methods
	SetNFTClass(nftClass dia.NFTClass) error

	// General methods
	GetKeys(table string) ([]string, error)
}

const (
	postgresKey = "postgres_credentials.txt"

	blockchainTable = "blockchain"
	nftclassTable   = "nftclass"
	nftTable        = "nft"

	// time format for blockchain genesis dates
	timeFormatBlockchain = "2006-01-02"
)

// RelDB is a relative database with redis caching layer.
type RelDB struct {
	URI            string
	postgresClient *pgx.Conn
	redisClient    *redis.Client
	pagesize       uint32
}

// NewRelDataStore returns a datastore with postgres client and redis cache.
func NewRelDataStore() (*RelDB, error) {
	log.Info("new rel datastore-------------------------")
	return NewRelDataStoreWithOptions(true, true)
}

// NewPostgresDataStore returns a datastore with postgres client and without redis caching layer.
func NewPostgresDataStore() (*RelDB, error) {
	return NewRelDataStoreWithOptions(true, false)
}

// NewCachingLayer returns a datastore with redis caching layer and without postgres client.
func NewCachingLayer() (*RelDB, error) {
	return NewRelDataStoreWithOptions(false, true)
}

// NewRelDataStoreWithOptions returns a postgres datastore and/or redis caching layer.
func NewRelDataStoreWithOptions(withPostgres bool, withRedis bool) (*RelDB, error) {
	var postgresClient *pgx.Conn
	var redisClient *redis.Client
	var err error
	// This environment variable is either set in docker-compose or empty
	executionMode := os.Getenv("EXEC_MODE")
	address := ""
	url := getPostgresURL(executionMode)
	if withPostgres {
		log.Info("connect to postgres server...")
		postgresClient, err = pgx.Connect(context.Background(), url)
		if err != nil {
			return nil, err
		}
		log.Info("...connection to postgres server established.")
	}
	if withRedis {
		// Run localhost for testing and server for production
		if executionMode == "production" {
			address = "redis:6379"
		} else {
			address = "localhost:6379"
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr:     address,
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		pong2, err := redisClient.Ping().Result()
		if err != nil {
			log.Error("NewDataStore redis", err)
		}
		log.Debug("NewDB", pong2)
	}
	return &RelDB{url, postgresClient, redisClient, 32}, nil
}

// GetKeys returns a slice of strings holding the names of the keys of @table in postgres
func (rdb *RelDB) GetKeys(table string) (keys []string, err error) {
	query := fmt.Sprintf("select column_name from information_schema.columns where table_name='%s'", table)
	rows, err := rdb.postgresClient.Query(context.Background(), query)
	if err != nil {
		return
	}
	for rows.Next() {
		val, err := rows.Values()
		if err != nil {
			return keys, err
		}
		keys = append(keys, val[0].(string))
	}
	return
}

func getPostgresURL(executionMode string) (url string) {
	if executionMode == "production" {
		url = "postgresql://postgres/postgres?user=postgres&password=" + getPostgresKeyFromSecrets(executionMode)
	} else {
		url = "postgresql://localhost/postgres?user=postgres&password=" + getPostgresKeyFromSecrets(executionMode)
	}
	return
}

func getPostgresKeyFromSecrets(executionMode string) string {
	var lines []string
	var file *os.File
	var err error
	if executionMode == "production" {
		pwd, _ := os.Getwd()
		log.Info("current directory: ", pwd)
		file, err = os.Open("/run/secrets/" + "postgres_credentials")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		gopath := os.Getenv("GOPATH")
		file, err = os.Open(gopath + "/src/github.com/diadata-org/diadata/secrets/" + postgresKey)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if len(lines) != 1 {
		log.Fatal("Secrets file should have exactly one line")
	}
	return lines[0]
}
