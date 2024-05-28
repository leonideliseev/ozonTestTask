package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/leonideliseev/ozonTestTask/graph"
	"github.com/leonideliseev/ozonTestTask/pkg/storage"

	"github.com/joho/godotenv"
	"github.com/leonideliseev/ozonTestTask/pkg/storage/in_memory"
	"github.com/leonideliseev/ozonTestTask/pkg/storage/postgresql"
	"github.com/sirupsen/logrus"
)

func main() {
	// загрузка файла .env если он есть
	if _, err := os.Stat(".env"); err == nil {
        err = godotenv.Load()
        if err != nil {
            log.Fatalf("Error loading .env file")
        }
		fmt.Println(".env file does loaded succefully")
    } else {
		fmt.Println(".env file does not exist")
	}

	// получение переменных окружения
	PORT := getEnv("APP_PORT", "8080")
	HOST_PORT := getEnv("HOST_PORT", "8080")
	dbStore := getEnv("DB_STORE", "false")

	var store storage.Storage
	var err error
	if dbStore == "true" { // подключение к бд
		connectionString := getEnv("DATABASE_URL", "postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable")
		store, err = postgresql.NewPostgreStore(connectionString)
		if err != nil {
			logrus.Fatalf("failed init db: %s", err.Error())
		}
	} else { // in-memory
		store = memory.NewInMemoryStore()
	}

	newResolver := graph.NewResolver(store)
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: newResolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", HOST_PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

// получение значения из окружения
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
		fmt.Printf("%s set from env\n", key)
        return value
    }
	fmt.Printf("%s set default\n", key)
    return defaultValue
}
