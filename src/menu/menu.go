package menu

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	// Redirige vers la page "/next"
	http.Redirect(w, r, "/next", http.StatusSeeOther)
}

// HomeHandler gère la page d'accueil
// Sert le fichier HTML templates.html
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// ServeFile envoie le fichier HTML au navigateur
	http.ServeFile(w, r, "./templates/templates.html")
}

// NextHandler gère la page suivante après redirection
// Affiche un message simple confirmant l'arrivée sur la page suivante
func NextHandler(w http.ResponseWriter, r *http.Request) {
	/* fmt.Fprintln(w, "Vous êtes sur la page suivante !") */
	http.ServeFile(w, r, "./templates/game.html")
}

// SetupRoutes configure toutes les routes du serveur web
// Cette fonction relie tous les handlers entre eux
func SetupRoutes() {
	// On crée un ServeMux local pour pouvoir appliquer un middleware de logging
	mux := http.NewServeMux()

	// Route pour servir les fichiers CSS et autres assets
	// FileServer permet de servir tous les fichiers du dossier assets/
	// On enregistre ce handler en priorité pour éviter que la route "/" ne l'emporte.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	// Route principale "/" → Page d'accueil avec le bouton
	mux.HandleFunc("/", HomeHandler)

	// Route "/welcome" → Gère le clic sur le bouton (redirection)
	mux.HandleFunc("/welcome", WelcomeHandler)

	// Route "/next" → Page de destination après redirection
	mux.HandleFunc("/next", NextHandler)

	// Route de debug pour lister le contenu du dossier assets
	mux.HandleFunc("/debug", DebugHandler)

	// Wrap du mux avec le logger puis on le passe au DefaultServeMux via Handle
	logged := loggingMiddleware(mux)
	http.Handle("/", logged)
}

// loggingMiddleware renvoie un Handler qui logge méthode, chemin, durée et code
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// wrapper pour capturer le status et la taille
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}
		start := time.Now()

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s -> %d (%d bytes) [%s]", r.Method, r.URL.Path, lrw.statusCode, lrw.size, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := l.ResponseWriter.Write(b)
	l.size += n
	return n, err
}

// DebugHandler liste le contenu du dossier ./assets pour debug
func DebugHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("./assets")
	if err != nil {
		http.Error(w, "Impossible de lire ./assets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "<html><body>")
	fmt.Fprintln(w, "<h1>Assets</h1>")
	fmt.Fprintln(w, "<ul>")
	for _, f := range files {
		info, _ := f.Info()
		name := f.Name()
		size := info.Size()
		fmt.Fprintf(w, "<li><a href=\"/assets/%s\">%s</a> — %d bytes</li>\n", name, name, size)
	}
	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "</body></html>")
}

// StartServer démarre le serveur web sur le port 8080
// Cette fonction lance le serveur et affiche un message de confirmation
func StartServer() {
	// Lis le port depuis la variable d'environnement PORT, sinon utilise 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	addr := ":" + port

	fmt.Printf("Serveur démarré sur le port %s...\n", port)
	fmt.Printf("Ouvrez votre navigateur sur : http://localhost:%s\n", port)

	// Démarre le serveur web (bloque le programme jusqu'à arrêt)
	log.Fatal(http.ListenAndServe(addr, nil))
}
