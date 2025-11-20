# Power4 Web

## Contexte & objectif
Adaptation web de Puissance 4 écrite en Go pour illustrer la gestion d’un état de jeu côté serveur et un rendu HTML minimaliste.  
L’application sert de support pédagogique : routing HTTP natif, templates `html/template` et interactions POST.  
Deux variantes (puissance-4 et puissance-5) et un mode bot basique montrent comment faire évoluer les règles sans changer l’UI.  
Le projet vise la simplicité de déploiement (un binaire Go) plutôt qu’une stack front moderne.

## Prérequis
- Go ≥ 1.25 (module `power4`)
- Navigateur moderne (testé sur Chrome/opéra GX)
- Optionnel : `air` ou `reflex` pour le rechargement à chaud

## Installation & exécution
```bash
git clone https://github.com/.../power4-web.git
cd power4-web
go mod tidy
go run ./...
```
Personnaliser le port :
```bash
set PORT=5000 && go run ./...      # Windows PowerShell/CMD
PORT=5000 go run ./...             # macOS/Linux
```
Build d’un binaire autonome :
```bash
go build -o power4.exe .
```

## Structure du dépôt
- `main.go` : point d’entrée qui configure le routing et lance le serveur.
- `src/menu/menu.go` : logique métier, état global de la partie et handlers HTTP.
- `templates/` : pages HTML (accueil, règles, variantes, modes de jeu).
- `static/styles.css` : styles globaux pour la grille et les boutons.
- `power4.exe` : binaire pré-compilé (peut être régénéré via `go build`).

## Fonctions clés
| Fonction | Signature | Rôle & dépendances | Complexité | Tests |
| --- | --- | --- | --- | --- |
| `SetupRoutes` | `func SetupRoutes()` | Déclare toutes les routes HTTP et sert les fichiers statiques via `net/http`. Dépend des handlers définis dans `menu.go`. | O(1) (initialisation). | Non automatisés (vérification manuelle via navigateur). |
| `StartServer` | `func StartServer()` | Lit `PORT`, initialise `math/rand` pour le bot et lance `http.ListenAndServe`. Dépend d’`os`, `log`, `net/http`. | Bloquante, coût minime hors écoute réseau. | Non, seulement tests manuels de démarrage. |
| `playHandler` | `func playHandler(w http.ResponseWriter, r *http.Request)` | Valide la requête POST, insère un pion, vérifie victoire/nul, pilote le bot. Dépend de `strconv`, `math/rand`, helpers `checkWin` et `isBoardFull`. | O(rows) pour trouver la première case libre + O(connectN) pour les contrôles. | Non, testé par scénario manuel (formulaire). |
| `checkWin` | `func checkWin(row, col, player int) bool` | Contrôle les 4 directions autour du dernier coup en fonction de `game.ConnectN`. Dépend uniquement de l’état global `game`. | O(connectN) par direction, constant dans la pratique. | Non, mais facilement testable via tests unitaires ciblant des grilles fictives. |
| `startHandler` / `startBotHandler` | `func startHandler(...)`, `func startBotHandler(...)` | Configurent la grille (6x7 ou 7x9), réinitialisent l’état et sélectionnent le mode `duo` ou `bot`. Utilisent `newBoard`. | O(rows*cols) pour recréer la matrice. | Non, couverture manuelle via boutons “Jouer” / “Bot”. |

## Décisions d’architecture & compromis techniques
L’application repose sur le serveur Go standard et des templates HTML pour éviter toute dépendance front-end. Cela permet un déploiement ultra-rapide dans un contexte académique, au prix d’une expérience utilisateur limitée (pas de mise à jour sans rechargement complet).
L’état de la partie est stocké dans une variable globale unique (`game`). Ce choix simplifie la logique mais interdit plusieurs parties simultanées et nécessite une isolation par session/mutex si l’on visait un usage multi-utilisateur.

## Qualité & tests
- Formatage via `gofmt` (intégré à Go).  
- Outils recommandés : `go vet ./...`, `staticcheck ./...`.  
- Pas de suite de tests dans le dépôt ; couverture empirique via sessions manuelles (objectif court terme : tests unitaires sur `checkWin`, `isBoardFull` et tests d’intégration HTTP grâce à `net/http/httptest`).

## Limites connues & améliorations
- Pas de persistance ni de sessions : une seule partie à la fois, rafraîchissement global.  
- Bot purement aléatoire ; aucune IA ni heuristique défensive.  
- Pas d’accessibilité spécifique (pas de support clavier/lecteurs d’écran).  
- Absence de pipeline CI/CD et de tests automatisés.  
Pistes : introduire un store par session (cookie ou UUID), ajouter une IA minimax simplifiée, convertir la vue en WebSocket ou htmx pour les tours de jeu, brancher GitHub Actions avec `go test`/`golangci-lint`.

## Crédits & licence
- Développement : équipe Power4 (Ynov).  
- Licence : usage académique interne (choisir MIT/BSD ou ajouter un fichier `LICENSE` avant publication publique).