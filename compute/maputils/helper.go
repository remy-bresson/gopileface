package maputils

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

func ExtractSomethingFromMap(queries map[string]string, what string, mandatory bool) (string, error) {
	extract := queries[what]
	if extract == "" {
		if mandatory {
			log.Error("empty ", what)
			return "", errors.New("empty " + what)
		}
	}
	return extract, nil
}
