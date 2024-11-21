package run

import (
	"context"
	"flag"
	"fmt"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"gochat/api"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func newServer() *gin.Engine {
	r := gin.Default()

	//if os.Getenv("ENV") == "production" {
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           os.Getenv("SENTRY_DSN"),
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for tracing.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	r.Use(sentrygin.New(sentrygin.Options{}))

	//}

	api.AddRoutes(r)
	return r
}
func main() {
	ctx := context.Background()
	if err := Run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func Run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	//Flags
	fs := flag.NewFlagSet("myflagset", flag.ExitOnError)
	var (
		port = fs.String("port", "8080", "Port to listen on")
	)
	err := fs.Parse(args[1:])
	if err != nil {
		log.Fatal(err)
	}

	srv := newServer()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: srv,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	// Create a wait group
	var wg sync.WaitGroup

	//Add to the counter
	wg.Add(1)

	go func() {
		// When the routine finishes, decrement the timer
		defer wg.Done()

		<-ctx.Done()
		// make a new context for the Shutdown (thanks Alessandro Rosetti)
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil

}
