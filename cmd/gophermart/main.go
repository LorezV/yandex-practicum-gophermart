package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/accural"
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/handlers"
	"github.com/LorezV/go-diploma.git/internal/middlewares"
	"github.com/LorezV/go-diploma.git/internal/repository/orderrepository"
	"github.com/LorezV/go-diploma.git/internal/repository/userrepository"
	"github.com/LorezV/go-diploma.git/internal/repository/withdrawalrepository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Config.RunAddress, "a", config.Config.RunAddress, "ip:port")
	flag.StringVar(&config.Config.DatabaseURI, "d", config.Config.DatabaseURI, "postgres://login:password@host:port/database")
	flag.StringVar(&config.Config.AccrualSystemAddress, "r", config.Config.AccrualSystemAddress, "An address of the Accrual System")
}

func main() {
	flag.Parse()

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := database.InitConnection(mainCtx); err != nil {
		panic(err)
	} else {
		fmt.Println("Connection to database was created.")
	}

	defer database.Connection.Close(mainCtx)

	if err := initRepositories(mainCtx); err != nil {
		fmt.Printf("Can't initialize repository: %T", err)
	} else {
		fmt.Println("Repositories was initialized.")
	}

	accural.InitAccrualClient()

	r := createRouter()

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		return http.ListenAndServe(config.Config.RunAddress, r)
	})

	g.Go(func() error {
		<-gCtx.Done()
		os.Exit(1)
		return nil
	})

	g.Go(func() error {
		return orderrepository.RunPollingStatuses(mainCtx)
	})

	err := g.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func createRouter() (r *chi.Mux) {
	r = chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.Authorization)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.Register)
		r.Post("/login", handlers.Login)
		r.Post("/orders", handlers.PostOrders)
		r.Get("/orders", handlers.GetOrders)
		r.Route("/balance", func(r chi.Router) {
			r.Get("/", handlers.GetBalance)
			r.Post("/withdraw", handlers.PostWithdraw)
		})
		r.Get("/withdrawals", handlers.GetWithdrawals)
	})

	return
}

func initRepositories(ctx context.Context) (err error) {
	err = userrepository.CreateUserTable(ctx)
	if err != nil {
		return
	}

	err = orderrepository.CreateOrderTable(ctx)
	if err != nil {
		return
	}

	err = withdrawalrepository.CreateWithdrawalTable(ctx)
	if err != nil {
		return
	}

	return nil
}
