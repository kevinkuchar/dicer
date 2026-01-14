package models

import "dicer/pkg/config"

/*************************************
* Ailments
*************************************/
type Ailments struct {
	Remaining []int
}

func CreateAilments(num int) *Ailments {
	slice := make([]int, num)

	for i := range slice {
		slice[i] = i + 1
	}

	ailments := &Ailments{Remaining: slice}
	return ailments
}

func (ailments *Ailments) HasAilments() bool {
	for a := range ailments.Remaining {
		if ailments.Remaining[a] != config.RemovedAilmentValue {
			return true
		}
	}
	return false
}

func (ailments *Ailments) HasAilment(num int) bool {
	if len(ailments.Remaining) < num || num < 1 {
		return false
	}
	return ailments.Remaining[num-1] != config.RemovedAilmentValue
}

func (ailments *Ailments) RemoveAilment(result int) {
	ailments.Remaining[result-1] = config.RemovedAilmentValue
}
