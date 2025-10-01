package main

import (
	"fmt"

	"github.com/IronWill79/blog-aggregator/internal/config"
)

func main() {
	cfg := config.Read()
	cfg.SetUser("IronWill79")
	cfg = config.Read()
	fmt.Println("Config file contents:")
	fmt.Printf("DB URL: %s\n", cfg.DBURL)
	fmt.Printf("Username: %s\n", cfg.Username)
}
