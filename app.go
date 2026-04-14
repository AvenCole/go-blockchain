package main

import (
	"context"

	gui "go-blockchain/internal/gui"
)

type App struct {
	service *gui.Service
}

func NewApp() *App {
	return &App{
		service: gui.NewService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.service.Startup(ctx)
}

func (a *App) Dashboard() (gui.DashboardData, error) {
	return a.service.Dashboard()
}

func (a *App) Wallets() ([]gui.WalletView, error) {
	return a.service.Wallets()
}

func (a *App) CreateWallet() (string, error) {
	return a.service.CreateWallet()
}

func (a *App) Blocks() ([]gui.BlockView, error) {
	return a.service.Blocks()
}

func (a *App) PendingTransactions() ([]string, error) {
	return a.service.PendingTransactions()
}

func (a *App) QueueTransaction(from, to string, amount int, fee int) (string, error) {
	return a.service.QueueTransaction(from, to, amount, fee)
}

func (a *App) MinePending(minerAddress string) (string, error) {
	return a.service.MinePending(minerAddress)
}
