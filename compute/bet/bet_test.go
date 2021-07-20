package main

import (
	"testing"
)

func TestBet_nominalPile(t *testing.T) {
	bet, err := checkBet("pile")
	if err != nil {
		t.Errorf("Check Bet not working properly, return an error where must not" + err.Error())
	}
	if bet != "pile" {
		t.Errorf("Check Bet not working properly")
	}
}

func TestBet_nominalFace(t *testing.T) {
	bet, err := checkBet("face")
	if err != nil {
		t.Errorf("Check Bet not working properly, return an error where must not" + err.Error())
	}
	if bet != "face" {
		t.Errorf("Check Bet not working properly")
	}
}
func TestBet_inError(t *testing.T) {
	_, err := checkBet("poil")
	if err == nil {
		t.Errorf("Check Bet not working properly, a error must have been returned")
	}
}
