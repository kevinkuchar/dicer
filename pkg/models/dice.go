package models

import "math/rand/v2"

/*************************************
* Dice
*************************************/
type Dice struct {
	Value int
}

func (d *Dice) Roll() {
	d.Value = rand.IntN(6) + 1
}

func CreateAndRollDie() Dice {
	die := Dice{}
	die.Roll()
	return die
}
