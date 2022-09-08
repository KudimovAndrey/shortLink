package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"
)

type service struct {
	storage Storage
}

const addr string = "localhost:8080"

type Storage interface {
	Put(string, string) error
	Get(string) (string, error)
	Close()
}

type InMemoryStorage struct {
	store map[string]string
}

func (i InMemoryStorage) Put(urlHash string, originalLink string) error {
	i.store[urlHash] = originalLink
	return nil
}

func (i InMemoryStorage) Get(urlHash string) (string, error) {
	return i.store[urlHash], nil

}

func (i InMemoryStorage) Close() {

}

type DbStorage struct {
	dbStore *pgx.Conn
}

func (d DbStorage) Put(urlHash string, originalLink string) error {
	_, err := d.dbStore.Exec(context.Background(), "INSERT INTO links (\"originalLink\", \"shortLink\") VALUES ($1,$2) "+
		"\tON CONFLICT (\"shortLink\") \tDO NOTHING;", originalLink, urlHash)
	if err != nil {
		return err
	}
	return nil
}

func (d DbStorage) Get(urlHash string) (string, error) {
	rows, err := d.dbStore.Query(context.Background(), ""+
		"SELECT \"originalLink\" "+
		"FROM links "+
		"WHERE \"shortLink\" = $1;", urlHash)
	if err != nil {
		return "", err
	}
	var originalLink string
	for rows.Next() {
		err = rows.Scan(&originalLink)
		if err != nil {
			return "", err
		}
	}
	return originalLink, nil
}

func (d DbStorage) Close() {
	d.dbStore.Close(context.Background())
}

func setupStorage() (Storage, error) {
	useDB := flag.Bool("d", false, "storage selection")
	flag.Parse()
	if *useDB {
		urlDb, err := getUrlDB()
		if err != nil {
			return nil, err
		}
		conn, err := pgx.Connect(context.Background(), urlDb)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			return nil, err
		}
		return DbStorage{conn}, nil
	}
	return InMemoryStorage{make(map[string]string)}, nil
}

func main() {
	mux := http.NewServeMux()
	storage, err := setupStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could't setup a storage %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()
	srv := service{storage: storage}
	mux.HandleFunc("/", srv.handle)
	http.ListenAndServe(addr, mux)
}

func (s *service) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.Post(w, r)
	case http.MethodGet:
		s.Get(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *service) Post(w http.ResponseWriter, r *http.Request) {
	var body string
	_, err := fmt.Fscanf(r.Body, "%s", &body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	urlHash := getHashURL(string(body))
	err = s.storage.Put(urlHash, string(body))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	shortLink := "http://" + addr + "/" + urlHash
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(shortLink))
}

func (s *service) Get(w http.ResponseWriter, r *http.Request) {
	urlHash := trimFirstRune(r.URL.Path)
	originalLink, err := s.storage.Get(urlHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(originalLink))
}

func getHashURL(url string) string {
	h := md5.New()
	h.Write([]byte(strings.ToLower(url)))
	return hex.EncodeToString(h.Sum(nil))
}

// todo something here looks not good
func getUrlDB() (string, error) {
	result := ""
	file, err := os.Open("linkFromDB.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result += scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return result, nil
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
