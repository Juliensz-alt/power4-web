package main

import (
	"fmt"
	"net/http"
)

def grille():
    # 6 lignes, 7 colonnes, initialisées à 0 (case vide)
    return [[0 for _ in range(7)] for _ in range(6)]