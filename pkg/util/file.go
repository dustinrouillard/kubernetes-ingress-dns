package util

import (
	"log"
	"os"
	"strings"
)

var HOSTS_PATH = Getenv("HOSTS_PATH", "/hosts")

func WriteHosts(hosts []string, file string) {
	log.Println("Ingresses changed, updating hosts file: " + file)

	text := []byte(strings.Join(hosts, "\n"))
	err := os.WriteFile(HOSTS_PATH+"/"+file, text, 0777)

	if err != nil {
		log.Println("Failed to write to hosts file", err)
	}
}
