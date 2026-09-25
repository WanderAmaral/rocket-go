package main

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

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

func toInt(s string) (int, error) {

	i, err := strconv.Atoi(s)

	if err != nil {
		return 0, errors.New(
			"não é permitido caractere diferente de número",
		)
	}

	return i, nil
}
