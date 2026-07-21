package utils

import (
	"log"
)

func InitProcessListeners() {
}

func RecoverHandler() {
	if r := recover(); r != nil {
		log.Printf("Recovered from panic: %v", r)
	}
}
