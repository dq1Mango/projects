package main

import (
	"fmt"
	"strconv"
)

func display(discard card, hands []hand, turn int) {

	colors := make(map[string]string)
	colors["none"] = "\033[0m"
	colors["black"] = "\033[0;40m"
	colors["red"] = "\033[0;31m"
	colors["yellow"] = "\033[0;33m"
	colors["green"] = "\033[0;32m"
	colors["blue"] = "\033[0;34m"

	fmt.Println("current card: ", colors[discard.color], discard, colors["none"])
	fmt.Print("your hand: ")
	for i := range 25 {
		cardToPrint := hands[turn].cards[i]
		if cardToPrint == nilCard {
			continue
		}
		fmt.Print(i+1, ": ", colors[cardToPrint.color], cardToPrint, colors["none"], ", ")
	}
	fmt.Println()
	fmt.Println("turn: ", turn)
}

func getInput() int64 {
	var input string
	for true {
		fmt.Print("Choose a card to play: ")
		fmt.Scan(&input)

		i, err := strconv.ParseInt(input, 10, 64)
		if err == nil && i >= 0 && i <= 25 {
			return i - 1
		}

		fmt.Println("INVALID INPUT (must be a number 0 - 25)")
	}
	return 0
}

func main() {
	deck := MakeDeck()
	deck.shuffle()

	players := 2
	hands := Deal(deck, players)

	discardPile := &Deck{
		cards: make([]card, 148),
		index: -1,
	}

	turnIndex := 0
	direction := 1

	newCard, err := deck.Pop()
	if err != nil {
		fmt.Println("could not get first card")
	}

	discardPile.Push(newCard)

	for !ValidStart(newCard) {
		newCard, err = deck.Pop()
		if err != nil {
			fmt.Println("could not get first card")
		}

		discardPile.Push(newCard)
	}

	for true {
		display(discardPile.Read(), hands, turnIndex)

		action := getInput()

		var cardToPlay card
		if action >= 0 && action < 25 {
			cardToPlay = hands[turnIndex].cards[action]
		} else if action == -1 {
			cardToPlay = nilCard
		} else {
			fmt.Println("shoulf have caugt this with prior check")
		}

		for !ValidMove(discardPile.Read(), cardToPlay) && action != -1 {
			fmt.Println("ILLEGAL MOVE ...")
			action := getInput()
			if action == 0 {
				continue
			}
			cardToPlay = hands[turnIndex].cards[action]
		}

		if action == -1 {
			drawCard, err := deck.Pop()
			if err != nil {
				fmt.Println("deck is out of cards most likely, should have caught this on the past turn")
			}
			err = hands[turnIndex].Add(drawCard)
			if err != nil {
				KillPlayer(hands, turnIndex, discardPile)
				players -= 1
			}

		} else {

			hands[turnIndex].Remove(action)
			if cardToPlay.color == "black" {
				if cardToPlay.number == 1 {
					// cry
				} else {
					if cardToPlay.number == 4 {
						direction *= -1
					}

					turnIndex = Abs(turnIndex+direction) % players

					err = Draw(cardToPlay.number, deck, &hands[turnIndex], discardPile)

					if err != nil {
						KillPlayer(hands, turnIndex, discardPile)
						players -= 1
					}
				}

			} else {
				// need to change this to implement 0s and 7s rule
				if cardToPlay.number < 10 {
					discardPile.Push(cardToPlay)
					// reverse
				} else if cardToPlay.number == 10 {
					direction *= -1
					// skip
				} else if cardToPlay.number == 11 {
					turnIndex = Abs(turnIndex+direction) % players
					// skip all
				} else if cardToPlay.number == 12 {
					turnIndex = Abs(turnIndex-direction) % players
					// draw 2
				} else if cardToPlay.number == 13 {

					turnIndex = Abs(turnIndex+direction) % players
					err = Draw(2, deck, &hands[turnIndex], discardPile)

					if err != nil {
						KillPlayer(hands, turnIndex, discardPile)
						players -= 1
					}
					// draw 4
				} else if cardToPlay.number == 14 {

					turnIndex = Abs(turnIndex+direction) % players
					err = Draw(4, deck, &hands[turnIndex], discardPile)

					if err != nil {
						KillPlayer(hands, turnIndex, discardPile)
						players -= 1
					}
				}

			}
		}
		//fmt.Println(hands[turnIndex].cards[action])
		turnIndex = Abs(turnIndex+direction) % players
		fmt.Println()
	}
}
