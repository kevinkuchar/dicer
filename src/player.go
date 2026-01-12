package main

/*************************************
* Player
*************************************/
type Player struct {
	Lives    int
	Ailments *Ailments
}

func CreatePlayer(numLives int, numAilments int) Player {
	var player *Player = &Player{Lives: numLives}
	player.Ailments = CreateAilments(numAilments)

	return *player
}

func (player *Player) HasLives() bool {
	return player.Lives > 0
}

func (player *Player) RemoveLife() {
	player.Lives--
}
