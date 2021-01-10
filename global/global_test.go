package global

import (
	"encoding/json"
	"testing"
)

func isEqual(a interface{}, b interface{}) bool {
	expect, _ := json.Marshal(a)
	got, _ := json.Marshal(b)

	return string(expect) == string(got)
}

func TestInit(t *testing.T) {

}
