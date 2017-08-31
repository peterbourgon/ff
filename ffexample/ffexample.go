package main

import (
	"log"
	"os"

	"github.com/peterbourgon/ff"
)

func main() {
	fs := ff.NewFlagSet("ffexample [flags]")
	var (
		addr = fs.String("addr", ":8080", "listen address")
		conf = fs.String("conf", "", "path to JSON config file (optional)")
	)
	if err := fs.Parse(os.Args[1:], ff.FromJSONFileVia("conf")); err != nil {
		log.Fatal(err)
	}
	log.Printf("args %v", os.Args)
	log.Printf("conf %q", *conf)
	log.Printf("addr %q", *addr)
}
