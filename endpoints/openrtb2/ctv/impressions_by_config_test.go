package ctv

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestCheck(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	//GetImpressionsByConfig(60, 90, 20, 45, 2, 10)
	//GetImpressionsByConfig(1, 90, 11, 15, 2, 8)
	//GetImpressionsByConfig(35, 65, 10, 35, 6, 40)
	//GetImpressionsByConfig(30, 90, 5, 15, 6, 15)
	//GetImpressionsByConfig(30, 60, 5, 9, 1, 6)
	//GetImpressionsByConfig(60, 90, 20, 45, 2, 5)
	GetImpressionsByConfig(88, 102, 30, 30, 3, 3)
}
