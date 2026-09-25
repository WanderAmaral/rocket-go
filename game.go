package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Orange  = "\033[38;5;208m"
	Pink    = "\033[38;5;198m"
	Blue    = "\033[32m"
	Magenta = "\033[35m"
	Reset   = "\033[0m"

	Underlined = "\033[4m"
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

	diretorio := "./quiz" // Subsitua pelo caminho da sua pasta

	entradas, err := os.ReadDir(diretorio)
	if err != nil {
		fmt.Printf("Erro ao ler o diretório: %v\n", err)
		return Theme{}, err
	}

	var themes []Theme

	for _, entrada := range entradas {
		if !entrada.IsDir() && strings.ToLower(filepath.Ext(entrada.Name())) == ".csv" {
			// Monta o caminho completo ex: "./quiz/Ciência e Tecnologia.csv"
			caminhoCompleto := filepath.Join(diretorio, entrada.Name())

			// Remove a extensão .csv do nome para exibir no menu
			nomeTema := strings.TrimSuffix(entrada.Name(), filepath.Ext(entrada.Name()))

			themes = append(themes, Theme{
				Name: nomeTema,
				File: caminhoCompleto, // Passa o caminho completo aqui!
			})
		}
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

// Essa função fica responsável por ler o teclado

func (g *GameState) Run() {
	if len(g.Questions) == 0 {
		fmt.Println("Nenhuma pergunta foi carregada.")
		return
	}

	input := make(chan string)

	var wrongAnswer int

	g.Points = 0
	g.CorrectAnswers = 0

	go readInput(input)

	// Embaralha as perguntas somente uma vez
	rand.Shuffle(len(g.Questions), func(i, j int) {
		g.Questions[i], g.Questions[j] =
			g.Questions[j], g.Questions[i]
	})

	// Percorre todas as perguntas
	for i, question := range g.Questions {
		fmt.Println()
		fmt.Println("----------------------------------")

		fmt.Printf(
			Yellow+"%d. %s\n"+Reset,
			i+1,
			question.Text,
		)

		// Mostra apenas as opções da pergunta atual
		for j, option := range question.Options {
			fmt.Printf("[%d] %s\n", j+1, option)
		}

		fmt.Println("Digite a alternativa:")
		fmt.Println("Você tem 10 segundos!")

		timer := time.NewTimer(10 * time.Second)
		ticker := time.NewTicker(1 * time.Second)

		tempoRestante := 10
		var answer int
		answered := false
		tempoEsgotado := false

		fmt.Printf("\rTempo restante: %02d segundos", tempoRestante)

		for !answered {
			select {
			case read, ok := <-input:
				if !ok {
					fmt.Println("\nEntrada encerrada.")
					ticker.Stop()
					timer.Stop()
					return
				}

				value, err := toInt(read)

				if err != nil {
					fmt.Printf("\n%s\n", err.Error())
					fmt.Println("Digite novamente:")
					continue
				}

				if value < 1 || value > len(question.Options) {
					fmt.Println("\nAlternativa inválida. Escolha uma opção válida.")
					continue
				}

				fmt.Printf("\nAlternativa escolhida: [%d]\n", value)

				answer = value
				answered = true

			case <-ticker.C:
				tempoRestante--

				if tempoRestante > 0 {
					fmt.Printf(
						"\rTempo restante: %02d segundos ",
						tempoRestante,
					)
				}

			case <-timer.C:
				fmt.Println("\n\n⏰ Tempo esgotado!")

				tempoEsgotado = true
				answered = true
			}
		}

		ticker.Stop()

		if !tempoEsgotado && !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		fmt.Println()

		if tempoEsgotado {
			wrongAnswer++

			fmt.Printf(
				"Resposta correta: [%d] %s\n",
				question.Answer,
				question.Options[question.Answer-1],
			)

		} else if answer == question.Answer {
			fmt.Println("✅ Parabéns, você acertou!")

			g.Points += 5
			g.CorrectAnswers++

		} else {
			fmt.Println("❌ Ops, você errou...")

			fmt.Printf(
				"Resposta correta: [%d] %s\n",
				question.Answer,
				question.Options[question.Answer-1],
			)

			wrongAnswer++
		}
	}

	fmt.Println()
	fmt.Println("==================================")

	if g.Points >= 10 {
		fmt.Println("Fim de jogo")
		fmt.Printf(
			"Parabéns você foi "+Underlined+Green+"aprovado,"+Reset+
				"%s! Você fez %d pontos.\n",
			g.Name,
			g.Points,
		)
	} else {
		fmt.Printf(
			"Você foi "+Underlined+Red+"reprovado,"+Reset+
				"%s! Você fez %d pontos.\n",
			g.Name,
			g.Points,
		)
	}

	total := len(g.Questions)

	fmt.Println("----------------------------------")
	fmt.Printf("Quantidade de perguntas: %d\n", total)
	fmt.Printf(
		Pink+"Respostas corretas: "+Reset+"%d\n",
		g.CorrectAnswers,
	)
	fmt.Printf(
		Orange+"Respostas erradas: "+Reset+"%d\n",
		wrongAnswer,
	)

	if total > 0 {
		utilization := float64(g.CorrectAnswers) /
			float64(total) * 100

		fmt.Printf(
			Magenta+"Aproveitamento: "+Reset+"%.0f%%\n",
			utilization,
		)
	}
}

func askAgain() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Deseja jogar novamente?")
		fmt.Println("Digite 's' para sim ou 'n' para não:")

		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		input = strings.TrimSpace(strings.ToLower(input))

		if input == "s" {
			return true
		}

		if input == "n" {
			return false
		}

		fmt.Println("Digite apenas 's' ou 'n'.")
	}
}
