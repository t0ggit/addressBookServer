package main

import (
	"addressBookServer/controllers/addressBookService"
	"addressBookServer/gates/psg"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	p, err := psg.NewPsg("postgres://127.0.0.1:5432/postgres", "postgres", "qwerty")
	if err != nil {
		log.Println("psg.NewPsg(): ", err)
	}

	abs := addressBookService.NewAddressBookService(":8080", p)

	signalCh := make(chan os.Signal, 1)     // канал для получения сигнала
	signal.Notify(signalCh, syscall.SIGINT) // привязываем его к сигналу SIGINT
	go func() {
		<-signalCh
		_ = abs.Close()
	}()

	abs.Start()
}
