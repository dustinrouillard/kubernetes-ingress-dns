package util

import (
	"log"
	"os"
	"strings"
)

var ValidModes = []string{"annotation", "class", "all"}
var ApplicationMode = strings.ToLower(Getenv("MODE", "all"))

func LoadEnv() {
	if !Contains(ValidModes, ApplicationMode) {
		log.Fatalln("error: Invalid configuration, unsupported MODE env supplied. Valid: [" + strings.Join(ValidModes, ", ") + "]")
		os.Exit(1)
	}

	if ApplicationMode != "annotation" && os.Getenv("INGRESS_CLASS") == "" {
		log.Println("warn: INGRESS_CLASS env is not set, defaulting to nginx")
	}
}
