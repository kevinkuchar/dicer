package main

/*************************************
* Ailments
*************************************/
type Ailments struct {
	remaining []int
}

func CreateAilments(num int) *Ailments {
	slice := make([]int, num)

	for i := range slice {
		slice[i] = i + 1
	}

	ailments := &Ailments{remaining: slice}
	return ailments
}

func (ailments *Ailments) HasAilments() bool {
	for a := range ailments.remaining {
		if ailments.remaining[a] != REMOVED_AILMENT_VALUE {
			return true
		}
	}
	return false
}

func (ailments *Ailments) HasAilment(num int) bool {
	if len(ailments.remaining) < num || num < 1 {
		return false
	}
	return ailments.remaining[num-1] != REMOVED_AILMENT_VALUE
}

func (ailments *Ailments) RemoveAilment(result int) {
	ailments.remaining[result-1] = REMOVED_AILMENT_VALUE
}
