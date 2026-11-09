package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

// logPath — куда пишется sqwave.log (пусто, если файл открыть не удалось).
var logPath string

// setupLogging открывает лог рядом с exe. Если каталог только для чтения —
// пробует рабочий каталог, затем temp. Возвращает функцию закрытия.
func setupLogging() func() {
	candidates := make([]string, 0, 3)
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "sqwave.log"))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "sqwave.log"))
	}
	candidates = append(candidates, filepath.Join(os.TempDir(), "sqwave.log"))

	for _, p := range candidates {
		flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		if os.Getenv("SQWAVE_RETRY") != "" {
			flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
		}
		f, err := os.OpenFile(p, flags, 0o644)
		if err != nil {
			continue
		}
		logPath = p
		log.SetOutput(io.MultiWriter(os.Stderr, f))
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		if os.Getenv("SQWAVE_RETRY") != "" {
			log.Printf("=== перезапуск (SQWAVE_RETRY), аргументы: %v ===", os.Args[1:])
		}
		log.Printf("Sqwave: старт, лог %s", p)
		return func() { _ = f.Close() }
	}
	return func() {}
}

// reportFatal пишет ошибку в лог и показывает её пользователю:
// в GUI-сборке под Windows консоли нет, без диалога ошибку не увидеть.
func reportFatal(title, msg string) {
	log.Printf("FATAL %s: %s", title, msg)
	text := msg
	if logPath != "" {
		text += "\n\nЛог: " + logPath
	}
	showErrorDialog("Sqwave — "+title, text)
}

// recoverPanic ловит панику в главной горутине (включая игровой цикл).
func recoverPanic() {
	if r := recover(); r != nil {
		reportFatal("падение", fmt.Sprintf("%v\n\n%s", r, debug.Stack()))
	}
}
