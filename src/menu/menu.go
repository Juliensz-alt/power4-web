package menu

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

// GameState représente l'état du jeu
type GameState struct {
	Board         [6][7]int `json:"board"`
	CurrentPlayer int       `json:"currentPlayer"`
	Winner        int       `json:"winner"`
	GameOver      bool      `json:"gameOver"`
	Message       string    `json:"message"`
	Started       bool      `json:"started"`
}

// GameData pour les templates
type GameData struct {
	GameState
	Player1Name string
	Player2Name string
}

var game = &GameState{
	Board:         [6][7]int{},
	CurrentPlayer: 1,
	Winner:        0,
	GameOver:      false,
	Message:       "Cliquez sur Jouer pour commencer !",
	Started:       false,
}

// SetupRoutes enregistre les handlers HTTP
func SetupRoutes() {
	// Serve the `static` folder at /static/ so templates can request /static/styles.css
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/quit", quitHandler)
	// Nouvelle routes: règles et page vide
	http.HandleFunc("/rules", rulesHandler)
	http.HandleFunc("/blank", blankHandler)
}

// StartServer démarre le serveur HTTP
func StartServer() {
	// Lire le port depuis la variable d'environnement PORT (par défaut 3000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Serveur Power4 démarré sur http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	funcMap := template.FuncMap{
		"seq": func(start, end int) []int {
			if end < start {
				return []int{}
			}
			s := make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				s = append(s, i)
			}
			return s
		},
	}

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Erreur lors du parsing du template: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	data := GameData{
		GameState:   *game,
		Player1Name: "Joueur 1",
		Player2Name: "Joueur 2",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
}

// rulesHandler affiche une page contenant les règles du jeu
func rulesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("rules.html").ParseFiles("templates/rules.html")
	if err != nil {
		log.Printf("Erreur lors du parsing du template rules: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// On peut fournir quelques infos de jeu si nécessaire
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template rules: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
}

// blankHandler affiche une page pour l'instant vide (placeholder)
func blankHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("blank.html").ParseFiles("templates/blank.html")
	if err != nil {
		log.Printf("Erreur lors du parsing du template blank: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template blank: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Empêcher de jouer tant que la partie n'a pas démarré
	if !game.Started {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	columnStr := r.FormValue("column")
	if columnStr == "" {
		game.Message = "Erreur: Aucune colonne sélectionnée"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	column, err := strconv.Atoi(columnStr)
	if err != nil || column < 1 || column > 7 {
		game.Message = "Erreur: Colonne invalide"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	columnIndex := column - 1

	if game.GameOver {
		game.Message = "Le jeu est terminé ! Cliquez sur 'Nouvelle partie' pour recommencer."
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	row := -1
	for i := 5; i >= 0; i-- {
		if game.Board[i][columnIndex] == 0 {
			row = i
			break
		}
	}

	if row == -1 {
		game.Message = "Erreur: Cette colonne est pleine !"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	game.Board[row][columnIndex] = game.CurrentPlayer

	if checkWin(row, columnIndex, game.CurrentPlayer) {
		game.Winner = game.CurrentPlayer
		game.GameOver = true
		game.Message = "🎉 Félicitations ! Joueur " + strconv.Itoa(game.CurrentPlayer) + " a gagné !"
	} else if isBoardFull() {
		game.GameOver = true
		game.Message = "🤝 Match nul ! La grille est pleine."
	} else {
		game.CurrentPlayer = 3 - game.CurrentPlayer
		game.Message = "Joueur " + strconv.Itoa(game.CurrentPlayer) + ", choisissez une colonne !"
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkWin(row, col, player int) bool {
	count := 1
	for i := col - 1; i >= 0 && game.Board[row][i] == player; i-- {
		count++
	}
	for i := col + 1; i < 7 && game.Board[row][i] == player; i++ {
		count++
	}
	if count >= 4 {
		return true
	}
	count = 1
	for i := row - 1; i >= 0 && game.Board[i][col] == player; i-- {
		count++
	}
	for i := row + 1; i < 6 && game.Board[i][col] == player; i++ {
		count++
	}
	if count >= 4 {
		return true
	}
	count = 1
	for i, j := row-1, col-1; i >= 0 && j >= 0 && game.Board[i][j] == player; i, j = i-1, j-1 {
		count++
	}
	for i, j := row+1, col+1; i < 6 && j < 7 && game.Board[i][j] == player; i, j = i+1, j+1 {
		count++
	}
	if count >= 4 {
		return true
	}
	count = 1
	for i, j := row-1, col+1; i >= 0 && j < 7 && game.Board[i][j] == player; i, j = i-1, j+1 {
		count++
	}
	for i, j := row+1, col-1; i < 6 && j >= 0 && game.Board[i][j] == player; i, j = i+1, j-1 {
		count++
	}
	return false
}

func isBoardFull() bool {
	for col := 0; col < 7; col++ {
		if game.Board[0][col] == 0 {
			return false
		}
	}
	return true
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	resetGame()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func resetGame() {
	game.Board = [6][7]int{}
	game.CurrentPlayer = 1
	game.Winner = 0
	game.GameOver = false
	// On conserve l'état Started pour rester dans la partie si on réinitialise
	if game.Started {
		game.Message = "Joueur 1, choisissez une colonne !"
	} else {
		game.Message = "Cliquez sur Jouer pour commencer !"
	}
}

// startHandler démarre une nouvelle partie et quitte le menu
func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Activer le mode jeu et reset propre
	game.Started = true
	game.Board = [6][7]int{}
	game.CurrentPlayer = 1
	game.Winner = 0
	game.GameOver = false
	game.Message = "Joueur 1, choisissez une colonne !"

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// quitHandler remet le jeu en mode menu (Started = false) et réinitialise l'état
func quitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	game.Started = false
	game.Board = [6][7]int{}
	game.CurrentPlayer = 1
	game.Winner = 0
	game.GameOver = false
	game.Message = "Cliquez sur Jouer pour commencer !"

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
