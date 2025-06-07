package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"post-comment-app/graph"
	"post-comment-app/storage"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	// Загружаем .env файл только при локальном запуске
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using system environment variables", "error", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	storageType := os.Getenv("STORAGE_TYPE")
	var store storage.Storage
	var pgStore *storage.PostgresStorage
	var err error

	switch storageType {
	case "postgres":
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatal("DATABASE_URL is required for postgres storage")
		}
		pgStore, err = storage.NewPostgresStorage(dsn)
		if err != nil {
			log.Fatalf("Failed to initialize postgres storage: %v", err)
		}
		store = pgStore
	default:
		slog.Info("Using in-memory storage")
		store = storage.NewInMemoryStorage()
	}

	if pgStore != nil {
		defer pgStore.Close()
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(store)}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
