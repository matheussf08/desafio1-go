package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Currency struct {
	UsdBrl Exchange
}

type Exchange struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
} 

type ExchangeDb struct {
	Bid string
}

func main() {
	http.HandleFunc("/cotacao", handlerExchange)
	http.ListenAndServe(":8080", nil)
}

func handlerExchange(w http.ResponseWriter, r *http.Request)  {
	db, err := sql.Open("sqlite3", "./exchange.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// // Create table
	// sqlStmt := `create table exchanges (id integer not null primary key autoincrement, bid text, create_date datetime default current_timestamp);`
	// _, err = db.Exec(sqlStmt)
	// if err != nil {
	// 	log.Printf("%q: %s\n", err, sqlStmt)
	// 	return
	// }

	ctx := context.Background()
	// ctx, cancel := context.WithTimeout(ctx, 200 * time.Millisecond)
	ctx, cancel := context.WithTimeout(ctx, 800 * time.Millisecond) // Criado com 800 milisegundos, porque com 200 está retornando deadline exceed
	defer cancel()

	exchange, error := getDollarExchange(ctx)

	if error != nil {
		msg := error.Error()
		// Verifica se o erro é do prazo do context que excedeu
		if strings.Contains(error.Error(), "context deadline exceeded") {
			msg = "context deadline exceed"
		}
		
		log.Print(msg)
		
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Insert bid into table
	if exchange != nil && exchange.Bid != "" {
		// ctx2, cancel := context.WithTimeout(ctx, 10 * time.Millisecond)
		ctx2, cancel := context.WithTimeout(ctx, 100 * time.Millisecond)
		defer cancel()
		
		stmt, err := db.Prepare("insert into exchanges(bid) values(?)")
		if err != nil {
			log.Print(error.Error())
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx2, exchange.Bid)
		if err !=  nil {
			msg := err.Error()
			// // Verifica se o erro é do prazo do context que excedeu
			if strings.Contains(err.Error(), "context deadline exceeded") {
				msg = "database context deadline exceed"
			}

			log.Print(msg)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exchange)
}

func getDollarExchange(ctx context.Context) (*Exchange, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	
	var c Currency
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return &c.UsdBrl, nil
}