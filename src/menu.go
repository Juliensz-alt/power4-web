package menu

import (
	"fmt"
	"net/http"
)


func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	// Redirige vers la page "/next"
	http.Redirect(w, r, "/next", http.StatusSeeOther)
}

// HomeHandler gère la page d'accueil
// Sert le fichier HTML templates.html
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// ServeFile envoie le fichier HTML au navigateur
	http.ServeFile(w, r, "templates/templates.html")
}

// NextHandler gère la page suivante après redirection
// Affiche un message simple confirmant l'arrivée sur la page suivante
func NextHandler(w http.ResponseWriter, r *http.Request) {
	/* fmt.Fprintln(w, "Vous êtes sur la page suivante !") */
	http.ServeFile(w, r, "templates/game.html")
}

// SetupRoutes configure toutes les routes du serveur web
// Cette fonction relie tous les handlers entre eux
func SetupRoutes() {
	// Route principale "/" → Page d'accueil avec le bouton
	http.HandleFunc("/", HomeHandler)
	
	// Route "/welcome" → Gère le clic sur le bouton (redirection)
	http.HandleFunc("/welcome", WelcomeHandler)
	
	// Route "/next" → Page de destination après redirection
	http.HandleFunc("/next", NextHandler)
	
	// Route pour servir les fichiers CSS et autres assets
	// FileServer permet de servir tous les fichiers du dossier assets/
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
}

// StartServer démarre le serveur web sur le port 8080
// Cette fonction lance le serveur et affiche un message de confirmation
func StartServer() {
	fmt.Println("Serveur démarré sur le port 8080...")
	fmt.Println("Ouvrez votre navigateur sur : http://localhost:8080")
	
	// Démarre le serveur web (bloque le programme jusqu'à arrêt)
	http.ListenAndServe(":8080", nil)
}