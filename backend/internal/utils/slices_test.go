package utils

import (
	"encoding/json"
	"testing"
)

func TestNaoNulo(t *testing.T) {
	var nulo []int
	b, _ := json.Marshal(NaoNulo(nulo))
	if string(b) != "[]" {
		t.Errorf("slice nil: got %s, quer []", b)
	}
	b, _ = json.Marshal(NaoNulo([]int{1, 2}))
	if string(b) != "[1,2]" {
		t.Errorf("slice com dados: got %s, quer [1,2]", b)
	}
}
