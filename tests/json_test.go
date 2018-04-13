package tests

import (
	"testing"
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"github.com/tidwall/gjson"
)

func TestJson_AllBNB(_t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	filename := dir+"/../data/exchangeInfo.json"
	fmt.Println(filename)

	b, err := ioutil.ReadFile(filename)
	value := gjson.GetBytes(b, `symbols.#[quoteAsset="BNB"]#.baseAsset`)
	fmt.Println(len(value.Array()))
}
