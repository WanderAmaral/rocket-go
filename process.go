package main

import (
	"encoding/csv"
	"errors"
	"os"
)

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
