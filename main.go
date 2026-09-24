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
	Name           string
	Points         int
	CorrectAnswers int
	Questions      []Question
}

type Theme struct {
	Name string
	File string
}

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)

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

func (t *GameState) EscolherTema() (Theme, error) {

	themes := []Theme{
		{Name: "Conhecimentos Gerais", File: "quizgo.csv"},
		{Name: "Ciência e Tecnologia", File: "quizgo_ciencia_tecnologia.csv"},
		{Name: "Esportes", File: "quizgo_esportes.csv"},
	}

	fmt.Printf(Red + "Escolha um tema.\n" + Reset)

	for index, theme := range themes {
		fmt.Printf(Green+"%d. %s\n"+Reset, index+1, theme.Name)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		opcaoTexto, err := reader.ReadString('\n')

		if err != nil {
			return Theme{}, errors.New("erro ao ler a escolha")
		}

		opcao, err := toInt(strings.TrimSpace(opcaoTexto))

		if err != nil {
			fmt.Println("Digite apenas um número.")
			continue
		}

		if opcao < 1 || opcao > len(themes) {
			fmt.Println("Tema inválido. Escolha uma opção disponível.")
			continue
		}

		temaSelecionado := themes[opcao-1]

		fmt.Printf("\nVocê escolheu o tema: %s!\n\n", temaSelecionado.Name)

		return temaSelecionado, nil
	}
}

func (g *GameState) ProcessCSV(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.New("Erro ao abrir arquivo CSV")
	}

	defer f.Close()

	reader := csv.NewReader(f)

	records, err := reader.ReadAll()
	if err != nil {
		return errors.New("Erro ao ler CSV")
	}

	for index, record := range records {

		// Ignora cabeçalho
		if index == 0 {
			continue
		}

		correctAnswer, err := toInt(record[len(record)-1])
		if err != nil {
			return errors.New("Resposta correta inválida no CSV")
		}

		question := Question{
			Text:    record[0],
			Options: record[1 : len(record)-1],
			Answer:  correctAnswer,
		}
		g.Questions = append(g.Questions, question)
	}
	return nil
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

	var wrongAnswer int

	// Goroutine responsável por ler o teclado
	go readInput(input)

	for i, question := range g.Questions {

		fmt.Println()
		fmt.Println("----------------------------------")

		fmt.Printf(
			Yellow+"%d. %s\n"+Reset,
			i+1,
			question.Text,
		)

		for j, option := range question.Options {
			fmt.Printf("[%d] %s\n", j+1, option)
		}

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
			//rightAnswer += 1
			g.CorrectAnswers++

		} else {

			fmt.Println("❌ Ops, você errou...")

			fmt.Printf(
				"Resposta correta: [%d] %s\n",
				question.Answer,
				question.Options[question.Answer-1],
			)
			wrongAnswer += 1
		}
	}

	fmt.Println()
	fmt.Println("==================================")

	if g.Points >= 10 {
		fmt.Println("Fim de jogo")
		fmt.Printf(
			"Parabéns você foi aprovado, %s! Você fez %d pontos.\n",
			g.Name,
			g.Points,
		)
	} else {
		fmt.Printf("Você foi reprovado, %s! com um total de %d pontos. \n", g.Name, g.Points)

	}

	total := len(g.Questions)

	fmt.Printf("Quantidade de perguntas: %d\n", total)
	fmt.Printf("Respostas corretas: %d\n", g.CorrectAnswers)
	fmt.Printf("Respostas erradas: %d\n", wrongAnswer)

	if total > 0 {
		utilization := float64(g.CorrectAnswers) / float64(total) * 100

		fmt.Printf("Aproveitamento: %.1f%%\n", utilization)
	}

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

	game.Init()
	temaEscolhido, err := game.EscolherTema()
	if err != nil {
		panic(err)
	}

	game.ProcessCSV(temaEscolhido.File)

	game.Run()
}
