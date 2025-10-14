package main

import (
	"power4/src/menu"
)

func main() {
	menu.SetupRoutes()
	menu.StartServer()
}
