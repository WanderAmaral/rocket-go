package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Question struct {
	Text    string
	Options []string
	Answer  int
}

type GameState struct {
	Name      string
	Points    int
	Questions []Question
}

func (g *GameState) Init() {
	fmt.Println("Seja bem-vindo(a) ao quiz")
	fmt.Println("Escreva o seu nome:")

	reader := bufio.NewReader(os.Stdin)

	name, err := reader.ReadString('\n')
	if err != nil {
		panic("Erro ao ler nome")
	}

	g.Name = strings.TrimSpace(name)

	fmt.Printf("\nVamos ao jogo, %s!\n\n", g.Name)
}

func (g *GameState) ProcessCSV() {
	f, err := os.Open("quizgo.csv")
	if err != nil {
		panic("Erro ao abrir arquivo CSV")
	}

	defer f.Close()

	reader := csv.NewReader(f)

	records, err := reader.ReadAll()
	if err != nil {
		panic("Erro ao ler CSV")
	}

	for index, record := range records {

		// Ignora cabeçalho
		if index == 0 {
			continue
		}

		correctAnswer, err := toInt(record[5])
		if err != nil {
			panic("Resposta correta inválida no CSV")
		}

		question := Question{
			Text:    record[0],
			Options: record[1:5],
			Answer:  correctAnswer,
		}

		g.Questions = append(g.Questions, question)
	}
}

// Essa função fica responsável por ler o teclado
func readInput(input chan string) {

	reader := bufio.NewReader(os.Stdin)

	for {

		text, err := reader.ReadString('\n')

		if err != nil {
			close(input)
			return
		}

		input <- strings.TrimSpace(text)
	}
}

func (g *GameState) Run() {

	// Canal que receberá as respostas do usuário
	input := make(chan string)

	// Goroutine responsável por ler o teclado
	go readInput(input)

	for i, question := range g.Questions {

		fmt.Println()
		fmt.Println("----------------------------------")

		fmt.Printf(
			"\033[33m%d. %s\033[0m\n",
			i+1,
			question.Text,
		)

		for j, option := range question.Options {
			fmt.Printf("[%d] %s\n", j+1, option)
		}

		fmt.Println()
		fmt.Println("Digite a alternativa:")
		fmt.Println("Você tem 10 segundos!")

		// Timer responsável pelo limite total
		timer := time.NewTimer(10 * time.Second)

		// Ticker dispara a cada 1 segundo
		ticker := time.NewTicker(1 * time.Second)

		tempoRestante := 10

		var answer int
		answered := false
		tempoEsgotado := false

		fmt.Printf("\rTempo restante: %02d segundos ", tempoRestante)

		for !answered {

			select {

			// Usuário respondeu
			case read, ok := <-input:

				if !ok {
					fmt.Println("\nEntrada encerrada.")
					return
				}

				value, err := toInt(read)

				if err != nil {
					fmt.Printf(
						"\n%s\n",
						err.Error(),
					)

					fmt.Println("Digite novamente:")

					continue
				}

				if value < 1 || value > len(question.Options) {
					fmt.Println(
						"\nAlternativa inválida. Escolha uma opção válida.",
					)

					continue
				}

				answer = value
				answered = true

			// Atualiza o cronômetro
			case <-ticker.C:

				tempoRestante--

				if tempoRestante > 0 {
					fmt.Printf(
						"\rTempo restante: %02d segundos ",
						tempoRestante,
					)
				}

			// 10 segundos acabaram
			case <-timer.C:

				fmt.Println("\n\n⏰ Tempo esgotado!")

				tempoEsgotado = true
				answered = true
			}
		}

		// Para o ticker
		ticker.Stop()

		// Para o timer caso o usuário tenha respondido antes
		if !tempoEsgotado {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}

		fmt.Println()

		if tempoEsgotado {

			fmt.Printf(
				"Resposta correta: [%d] %s\n",
				question.Answer,
				question.Options[question.Answer-1],
			)

		} else if answer == question.Answer {

			fmt.Println("✅ Parabéns, você acertou!")

			g.Points += 5

		} else {

			fmt.Println("❌ Ops, você errou...")

			fmt.Printf(
				"Resposta correta: [%d] %s\n",
				question.Answer,
				question.Options[question.Answer-1],
			)
		}
	}

	fmt.Println()
	fmt.Println("==================================")

	fmt.Printf(
		"Fim de jogo, %s! Você fez %d pontos.\n",
		g.Name,
		g.Points,
	)
}

func toInt(s string) (int, error) {

	i, err := strconv.Atoi(s)

	if err != nil {
		return 0, errors.New(
			"não é permitido caractere diferente de número",
		)
	}

	return i, nil
}

func main() {

	game := &GameState{
		Points: 0,
	}

	go game.ProcessCSV()

	game.Init()

	game.Run()
}