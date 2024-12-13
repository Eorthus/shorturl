package profiler

import (
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var profileFile = flag.String("profile", "base.pprof", "Name of the profile output file")

func StartProfiling() func() {
	// Создаем директорию для профилей если её нет
	if err := os.MkdirAll("profiles", 0755); err != nil {
		log.Fatal(err)
	}

	// Используем значение из флага для имени файла
	f, err := os.Create("profiles/" + *profileFile)
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем сборку мусора перед получением профиля памяти
	runtime.GC()

	// Записываем профиль памяти
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal(err)
	}

	// Возвращаем функцию cleanup
	return func() {
		f.Close()
	}
}
