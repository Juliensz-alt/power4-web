package main

import (
	"github.com/elsaCzeyn/testpower4/src/menu" // Import avec le bon nom de module
)

func main() {
	// Configuration de toutes les routes web
	// Cette fonction relie tous les handlers entre eux
	menu.SetupRoutes()
	
	// Démarrage du serveur web
	// Cette fonction lance le serveur sur le port 8080
	menu.StartServer()
}
