package utils

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

func ExtractSomethingFromMap(queries map[string]string, what string) (string, error) {
	extract := queries[what]
	if extract == "" {
		log.Error("empty", what)
		return "", errors.New("empty" + what)

	}
	return extract, nil
}
