package helpers

import (
	"os"
	"github.com/sirupsen/logrus"
	"strings"
)

func WorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	index := strings.Index(dir, "TradeBot")
	path := dir[:index] + "TradeBot"
	return path
}
