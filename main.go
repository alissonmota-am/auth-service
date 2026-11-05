package main

import (
	"context"
	"database/sql"
	// "fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// App struct (para injeção de dependência)
type App struct {
	DB        *sql.DB
	MasterKey string

	// Injecao de falha (demo de self-healing). O estado vive so na memoria do
	// processo: um pod novo sempre nasce com failing = false, por isso um
	// "kubectl rollout restart" resolve a falha.
	chaosEnabled bool
	failing      atomic.Bool
}

func main() {
	// Carrega o .env para desenvolvimento local. Em produção, isso não fará nada.
	_ = godotenv.Load()

	// --- OpenTelemetry ---
	ctx := context.Background()
	shutdownTelemetry := initTelemetry(ctx)
	defer func() { _ = shutdownTelemetry(ctx) }()

	// --- Configuração ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001" // Porta padrão
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL deve ser definida")
	}

	masterKey := os.Getenv("MASTER_KEY")
	if masterKey == "" {
		log.Fatal("MASTER_KEY deve ser definida")
	}

	// --- Conexão com o Banco ---
	db, err := connectDB(databaseURL)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	app := &App{
		DB:           db,
		MasterKey:    masterKey,
		chaosEnabled: os.Getenv("CHAOS_ENABLED") == "true",
	}

	// --- Rotas da API ---
	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthHandler)

	// Endpoint público para validar uma chave
	mux.HandleFunc("/validate", app.validateKeyHandler)

	// Endpoints de "admin" para criar/gerenciar chaves
	// Eles são protegidos pelo middleware de autenticação
	mux.Handle("/admin/keys", app.masterKeyAuthMiddleware(http.HandlerFunc(app.createKeyHandler)))

	// Injecao de falha para a demo de self-healing. A rota so existe com
	// CHAOS_ENABLED=true e exige a MASTER_KEY.
	if app.chaosEnabled {
		mux.Handle("/admin/chaos", app.masterKeyAuthMiddleware(http.HandlerFunc(app.chaosHandler)))
		log.Println("[CHAOS] Injecao de falha HABILITADA em /admin/chaos")
	}

	// Envolve o mux com otelhttp — cria um span automatico para cada requisicao HTTP
	handler := otelhttp.NewHandler(mux, "auth-service")

	log.Printf("Serviço de Autenticação (Go) rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

// connectDB inicializa e testa a conexão com o PostgreSQL
func connectDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Conectado ao PostgreSQL com sucesso!")
	return db, nil
}