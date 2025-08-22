package main

import (
	"fmt"
	"log/slog"

	"github.com/Serajian/cache-db.git/database"
)

func main() {
	db := database.NewDatabase()
	db.Set("k1", 23)
	db.Set("k2", 62)

	err := db.Persist("database.gob")
	if err != nil {
		slog.Error("err persist", err)
		return
	}
	//err = db.Load("database.gob")
	//if err != nil {
	//	slog.Error("err load", err)
	//	return
	//}
	value, ok1 := db.Get("k1")
	fmt.Println(value, ok1)
	value2, ok2 := db.Get("k2")
	fmt.Println(value2, ok2)

}
