package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/accural"
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/handlers"
	"github.com/LorezV/go-diploma.git/internal/middlewares"
	"github.com/LorezV/go-diploma.git/internal/repositories/orderrepository"
	"github.com/LorezV/go-diploma.git/internal/repositories/userrepository"
	"github.com/LorezV/go-diploma.git/internal/repositories/withdrawalrepository"
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

func Init() {
	rand.Seed(time.Now().UnixNano())

	err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Config.RunAddress, "a", config.Config.RunAddress, "ip:port")
	flag.StringVar(&config.Config.DatabaseURI, "d", config.Config.DatabaseURI, "postgres://login:password@host:port/database")
	flag.StringVar(&config.Config.AccrualSystemAddress, "r", config.Config.AccrualSystemAddress, "ip:port")
}

func Run() {
	flag.Parse()

	if err := database.InitConnection(); err != nil {
		panic(err)
	} else {
		fmt.Println("Connection to database was created.")
	}

	defer database.Connection.Close(context.Background())

	if err := initRepositories(); err != nil {
		fmt.Printf("Can't initialize repositories: %T", err)
	} else {
		fmt.Println("Repositories was initialized.")
	}

	accural.AccrualClient = accural.MakeAccrualClient()

	r := createRouter()

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		return http.ListenAndServe(config.Config.RunAddress, r)
	})

	g.Go(func() error {
		<-gCtx.Done()
		os.Exit(1)
		return errors.New("application shutdown")
	})

	g.Go(func() error {
		if err := orderrepository.RunPollingStatuses(mainCtx); err != nil {
			if errors.Is(err, accural.ErrAccrualSystemNoContent) {
				return nil
			}

			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Println(err.Error())
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

func initRepositories() (err error) {
	err = userrepository.CreateUserTable(context.Background())
	if err != nil {
		return
	}

	err = orderrepository.CreateOrderTable(context.Background())
	if err != nil {
		return
	}

	err = withdrawalrepository.CreateWithdrawalTable(context.Background())
	if err != nil {
		return
	}

	return nil
}
