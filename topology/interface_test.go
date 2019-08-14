package topology

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDecompose(t *testing.T) {
	data, _ := ioutil.ReadFile("./test-examples/00-simple.yaml")

	res := Decompose(string(data))

	fmt.Printf("%v", res.Exchanges[0])
}
