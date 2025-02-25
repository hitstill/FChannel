package config

import (
	"log"
	"os"
)

var Log = log.New(os.Stdout, "", log.Ltime)
