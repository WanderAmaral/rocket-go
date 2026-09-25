package main

func main() {

	game := &GameState{
		Points: 0,
	}

	game.Init()
	temaEscolhido, err := game.EscolherTema()
	if err != nil {
		panic(err)
	}

	err = game.ProcessCSV(temaEscolhido.File)

	if err != nil {
		panic(err)
	}

	game.Run()
}
