package main

type Question struct {
	Text    string
	Options []string
	Answer  int
}

type Theme struct {
	Name string
	File string
}

type GameState struct {
	Name           string
	Points         int
	CorrectAnswers int
	Questions      []Question
}
