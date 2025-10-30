package menu

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// GameState représente l'état du jeu
type GameState struct {
	Board         [][]int `json:"board"`
	Rows          int     `json:"rows"`
	Cols          int     `json:"cols"`
	ConnectN      int     `json:"connectN"`
	CurrentPlayer int     `json:"currentPlayer"`
	Winner        int     `json:"winner"`
	GameOver      bool    `json:"gameOver"`
	Message       string  `json:"message"`
	Started       bool    `json:"started"`
	Mode          string  `json:"mode"` // "duo" or "bot"
}

// GameData pour les templates
type GameData struct {
	GameState
	Player1Name string
	Player2Name string
	CSSVersion  int64
}

func newBoard(rows, cols int) [][]int {
	b := make([][]int, rows)
	for i := range b {
		b[i] = make([]int, cols)
	}
	return b
}

var game = &GameState{
	Rows:          6,
	Cols:          7,
	ConnectN:      4,
	Board:         newBoard(6, 7),
	CurrentPlayer: 1,
	Winner:        0,
	GameOver:      false,
	Message:       "Cliquez sur Jouer pour commencer !",
	Started:       false,
	Mode:          "",
}

// mutex pour protéger l'accès concurrent à game (pour le bot async)
// (no mutex needed for synchronous bot updates)

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
	http.HandleFunc("/variant", variantHandler)
	// Mode selection: duo or bot
	http.HandleFunc("/game-mode", gameModeHandler)
	http.HandleFunc("/start-bot", startBotHandler)
}

// variantHandler affiche la page qui permet de choisir la variante (4 ou 5)
func variantHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("variant.html").ParseFiles("templates/variant.html")
	if err != nil {
		log.Printf("Erreur lors du parsing du template variant: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	data := struct{ CSSVersion int64 }{CSSVersion: time.Now().Unix()}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template variant: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
}

// StartServer démarre le serveur HTTP
func StartServer() {
	// Lire le port depuis la variable d'environnement PORT (par défaut 3000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Initialiser le générateur aléatoire pour le bot
	rand.Seed(time.Now().UnixNano())

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
		CSSVersion:  time.Now().Unix(),
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
	if err != nil || column < 1 || column > game.Cols {
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
	for i := game.Rows - 1; i >= 0; i-- {
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

	// Appliquer le coup du joueur
	game.Board[row][columnIndex] = game.CurrentPlayer

	if checkWin(row, columnIndex, game.CurrentPlayer) {
		game.Winner = game.CurrentPlayer
		game.GameOver = true
		game.Message = "🎉 Félicitations ! Joueur " + strconv.Itoa(game.CurrentPlayer) + " a gagné !"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if isBoardFull() {
		game.GameOver = true
		game.Message = "🤝 Match nul ! La grille est pleine."
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Passage au joueur suivant
	game.CurrentPlayer = 3 - game.CurrentPlayer
	game.Message = "Joueur " + strconv.Itoa(game.CurrentPlayer) + ", choisissez une colonne !"

	// Si mode bot et c'est au tour du bot (joueur 2), le bot joue immédiatement (synchrones)
	if game.Mode == "bot" && game.CurrentPlayer == 2 && !game.GameOver {
		// choisir une colonne aléatoire parmi celles qui ne sont pas pleines
		available := make([]int, 0, game.Cols)
		for c := 0; c < 7; c++ {
			if c < game.Cols && game.Board[0][c] == 0 {
				available = append(available, c)
			}
		}
		if len(available) > 0 {
			c := available[rand.Intn(len(available))]
			rIndex := -1
			for i := game.Rows - 1; i >= 0; i-- {
				if game.Board[i][c] == 0 {
					rIndex = i
					break
				}
			}
			if rIndex != -1 {
				game.Board[rIndex][c] = 2
				if checkWin(rIndex, c, 2) {
					game.Winner = 2
					game.GameOver = true
					game.Message = "Le bot a gagné !"
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
				if isBoardFull() {
					game.GameOver = true
					game.Message = "🤝 Match nul ! La grille est pleine."
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
				// revenir au joueur 1
				game.CurrentPlayer = 1
				game.Message = "Joueur 1, choisissez une colonne !"
			}
		} else {
			game.GameOver = true
			game.Message = "🤝 Match nul ! La grille est pleine."
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkWin(row, col, player int) bool {
	// Check directions: horizontal, vertical, diag (NW-SE), diag (NE-SW)
	// Use game.ConnectN as the needed aligned tokens count.

	need := game.ConnectN

	// helper to count in a direction (dr, dc), excluding the starting cell
	countDir := func(dr, dc int) int {
		cnt := 0
		r, c := row+dr, col+dc
		for r >= 0 && r < game.Rows && c >= 0 && c < game.Cols && game.Board[r][c] == player {
			cnt++
			r += dr
			c += dc
		}
		return cnt
	}

	// horizontal
	if 1+countDir(0, -1)+countDir(0, 1) >= need {
		return true
	}
	// vertical
	if 1+countDir(-1, 0)+countDir(1, 0) >= need {
		return true
	}
	// diagonal NW-SE
	if 1+countDir(-1, -1)+countDir(1, 1) >= need {
		return true
	}
	// diagonal NE-SW
	if 1+countDir(-1, 1)+countDir(1, -1) >= need {
		return true
	}
	return false
}

func isBoardFull() bool {
	for col := 0; col < game.Cols; col++ {
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
	game.Board = newBoard(game.Rows, game.Cols)
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

	// Read optional variant (4 or 5) from the form and configure the board
	variant := r.FormValue("variant")
	if variant == "5" {
		game.Rows = 7
		game.Cols = 9
		game.ConnectN = 5
	} else {
		game.Rows = 6
		game.Cols = 7
		game.ConnectN = 4
	}
	// Activer le mode jeu et reset propre
	game.Started = true
	game.Board = newBoard(game.Rows, game.Cols)
	game.CurrentPlayer = 1
	game.Winner = 0
	game.GameOver = false
	game.Message = "Joueur 1, choisissez une colonne !"

	// Par défaut le start normal est en duo
	game.Mode = "duo"

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// quitHandler remet le jeu en mode menu (Started = false) et réinitialise l'état
func quitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	game.Started = false
	// Reset to default variant (Puissance 4) when quitting to menu
	game.Rows = 6
	game.Cols = 7
	game.ConnectN = 4
	game.Board = newBoard(game.Rows, game.Cols)
	game.CurrentPlayer = 1
	game.Winner = 0
	game.GameOver = false
	game.Message = "Cliquez sur Jouer pour commencer !"
	game.Mode = ""

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// gameModeHandler affiche la page qui permet de choisir Duo ou Bot
func gameModeHandler(w http.ResponseWriter, r *http.Request) {
	// This handler accepts POST from the variant selector with form "variant"
	// and renders the mode selection page (duo or bot) with that variant.
	if r.Method != "POST" && r.Method != "GET" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var variantInt int
	if r.Method == "POST" {
		if err := r.ParseForm(); err == nil {
			v := r.FormValue("variant")
			if v == "5" {
				variantInt = 5
			} else {
				variantInt = 4
			}
		}
	}

	tmpl, err := template.New("game_mode.html").ParseFiles("templates/game_mode.html")
	if err != nil {
		log.Printf("Erreur lors du parsing du template game_mode: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	data := struct{
		Variant    int
		CSSVersion int64
	}{Variant: variantInt, CSSVersion: time.Now().Unix()}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template game_mode: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
}

// startBotHandler démarre une partie contre le bot (bot = joueur 2)
func startBotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Read variant if any
	variant := r.FormValue("variant")
	if variant == "5" {
		game.Rows = 7
		game.Cols = 9
		game.ConnectN = 5
	} else {
		game.Rows = 6
		game.Cols = 7
		game.ConnectN = 4
	}

	game.Started = true
	game.Board = newBoard(game.Rows, game.Cols)
	game.CurrentPlayer = 1
	game.Winner = 0
	game.GameOver = false
	game.Message = "Joueur 1, choisissez une colonne !"
	game.Mode = "bot"

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
