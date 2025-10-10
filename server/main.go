package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-mservice-template/server/config"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-mservice-template/server/grpc"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-mservice-template/server/http"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-mservice-template/server/log"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-mservice-template/server/repository"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-user-mservice/common/utils"
)

var (
	sigCh = make(chan os.Signal, 1)
)

func main() {
	config.Init()
	log.InitLogger()
	repository.Init()
	utils.InitJwtSecret()
	go grpc.Init(sigCh)
	go http.Init(sigCh)
	// listen terminage signal
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh // Block until signal is received
	log.Logger.Infof("Received signal: %v, shutting down...", sig)
}
