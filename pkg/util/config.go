package util

import (
	"log"
	"os"
	"strings"
)

var ValidModes = []string{"annotation", "all"}
var ApplicationMode = strings.ToLower(Getenv("MODE", "annotation"))

func LoadEnv() {
	if !Contains(ValidModes, ApplicationMode) {
		log.Fatalln("error: Invalid configuration, unsupported MODE env supplied. Valid: [" + strings.Join(ValidModes, ", ") + "]")
		os.Exit(1)
	}
}
