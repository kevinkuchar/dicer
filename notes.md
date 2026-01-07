# Game Dicer

## First Iteration

### Game Flow
- Single Player
- Player has 3 hearts
- Player starts with ailments numbered 1-9
- Player rolls 3 6-sided dice
- Player can choose to reroll 1-3 dice
- Player must create an expression using + - * / and the 3 numbers (example 3 + 2 - 1)
- Order of operations matters
- The result of the expression must equal one of the remaining ailments
- Players heals the ailment equal to the result
- If player cannot heal an ailment player loses a life
- Player rolls again
- Continue until all ailments are healed or lives are lost
- If all ailments are healed - victory
- If all hearts are lost - defeat

### Todo
[x] Create Dice Structure
[x] Add ability to roll a dice
[x] Add ability to create three dice and roll them
[ ] Validate the infix expression contains only numbers you rolled
[ ] Refactor Roll struct to be a turn struct to reflect the logic
[ ] Apply result of roll to ailments and lives
[ ] Show game finished screen
[ ] Color Ailments Red when complete
[ ] Color Lives Red