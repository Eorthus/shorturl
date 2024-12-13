package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
)

var profileFile = flag.String("profile", "base.pprof", "Name of the profile output file")

func main() {
	flag.Parse()

	// Создаем директорию для профилей если её нет
	if err := os.MkdirAll("profiles", 0755); err != nil {
		log.Fatal(err)
	}

	// Используем значение из флага для имени файла
	f, err := os.Create("profiles/" + *profileFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Запускаем HTTP сервер для профилирования
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Запускаем сборку мусора перед получением профиля памяти
	runtime.GC()

	// Записываем профиль памяти
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal(err)
	}

	log.Printf("Profile saved to profiles/%s\n", *profileFile)
}
