package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"chmethod_lab2/internal/webui"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "адрес локального веб-сервера")
	flag.Parse()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v. Выберите другой порт: -addr 127.0.0.1:8081", err)
	}
	server := &http.Server{Handler: webui.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	fmt.Printf("Численные методы · ЛР2 · Вариант 9\nОткройте http://%s\nCtrl+C — завершить работу.\n", listener.Addr())
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
