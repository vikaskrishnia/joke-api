package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

// Joke represents a joke object with content in multiple languages
type Joke struct {
	ID       int               `json:"id"`
	Content  map[string]string `json:"content"`
	Category string            `json:"category"`
}

// JokeResponse is the structure returned by the API
type JokeResponse struct {
	Joke     string `json:"joke"`
	Language string `json:"language"`
}

// ErrorResponse represents an error message
type ErrorResponse struct {
	Error string `json:"error"`
}

// Global variables
var (
	jokes = []Joke{
		{
			ID: 1,
			Content: map[string]string{
				"en": "Why don't scientists trust atoms? Because they make up everything.",
				"es": "¿Por qué los científicos no confían en los átomos? Porque lo componen todo.",
				"fr": "Pourquoi les scientifiques ne font-ils pas confiance aux atomes ? Parce qu'ils inventent tout.",
				"de": "Warum vertrauen Wissenschaftler Atomen nicht? Weil sie alles erfinden.",
				"hi": "वैज्ञानिक परमाणुओं पर विश्वास क्यों नहीं करते? क्योंकि वे सब कुछ बना देते हैं।",
			},
			Category: "science",
		},
		{
			ID: 2,
			Content: map[string]string{
				"en": "I told my wife she was drawing her eyebrows too high. She looked surprised.",
				"es": "Le dije a mi esposa que estaba dibujando sus cejas demasiado altas. Parecía sorprendida.",
				"fr": "J'ai dit à ma femme qu'elle dessinait ses sourcils trop haut. Elle avait l'air surprise.",
				"de": "Ich sagte meiner Frau, dass sie ihre Augenbrauen zu hoch zeichnet. Sie sah überrascht aus.",
				"hi": "मैंने अपनी पत्नी से कहा कि वह अपनी भौंहें बहुत ऊंची बना रही है। वह आश्चर्यचकित दिखीं।",
			},
			Category: "pun",
		},
		{
			ID: 3,
			Content: map[string]string{
				"en": "Why did the programmer go broke? Because he lost his domain in a crash.",
				"es": "¿Por qué el programador se quedó sin dinero? Porque perdió su dominio en un accidente.",
				"fr": "Pourquoi le programmeur est-il devenu pauvre ? Parce qu'il a perdu son domaine dans un crash.",
				"de": "Warum ging der Programmierer pleite? Weil er seine Domain bei einem Absturz verloren hat.",
				"hi": "प्रोग्रामर कंगाल क्यों हो गया? क्योंकि उसने क्रैश में अपना डोमेन खो दिया।",
			},
			Category: "programming",
		},
		{
			ID: 4,
			Content: map[string]string{
				"en": "Why don't programmers like nature? It has too many bugs.",
				"es": "¿Por qué a los programadores no les gusta la naturaleza? Tiene demasiados insectos.",
				"fr": "Pourquoi les programmeurs n'aiment pas la nature ? Elle a trop de bugs.",
				"de": "Warum mögen Programmierer die Natur nicht? Sie hat zu viele Bugs.",
				"hi": "प्रोग्रामर प्रकृति को क्यों पसंद नहीं करते? इसमें बहुत सारे बग हैं।",
			},
			Category: "programming",
		},
		{
			ID: 5,
			Content: map[string]string{
				"en": "What do you call a fake noodle? An impasta.",
				"es": "¿Cómo se llama un fideo falso? Un impasta.",
				"fr": "Comment appelle-t-on de fausses nouilles ? Des impasstas.",
				"de": "Wie nennt man eine gefälschte Nudel? Eine Impasta.",
				"hi": "नकली नूडल्स को क्या कहते हैं? इम्पास्ता।",
			},
			Category: "food",
		},
	}

	// Metrics
	meter                 api.Meter
	requestCounter        api.Int64Counter
	responseTimeHistogram api.Float64Histogram
)

func initMetrics() *prometheus.Exporter {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	meter = provider.Meter("joke-api")

	// Create a counter for requests
	requestCounter, _ = meter.Int64Counter(
		"api.requests.total",
		api.WithDescription("Total number of API requests"),
	)

	// Create a histogram for response times
	responseTimeHistogram, _ = meter.Float64Histogram(
		"api.response.time",
		api.WithDescription("API response time in seconds"),
		api.WithUnit("s"),
	)

	return exporter
}

// Middleware to record metrics
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create attributes
		attrs := []attribute.KeyValue{
			attribute.String("endpoint", r.URL.Path),
			attribute.String("method", r.Method),
		}

		// Count the request
		requestCounter.Add(context.Background(), 1, api.WithAttributes(attrs...))

		// Call the next handler
		next.ServeHTTP(w, r)

		// Record response time
		elapsedTime := time.Since(startTime).Seconds()
		responseTimeAttrs := append(attrs, attribute.String("status_code", "200")) // In a real app, capture actual status code
		responseTimeHistogram.Record(context.Background(), elapsedTime, api.WithAttributes(responseTimeAttrs...))
	})
}

// getRandomJoke returns a random joke in the requested language
func getRandomJoke(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en" // Default to English
	}

	// Check if language is supported
	supportedLangs := map[string]bool{
		"en": true,
		"es": true,
		"fr": true,
		"de": true,
		"hi": true, // Added Hindi support
	}

	if !supportedLangs[lang] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unsupported language. Use: en, es, fr, de, or hi."})
		return
	}

	// Get a random joke
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	randomIndex := randGenerator.Intn(len(jokes))
	randomJoke := jokes[randomIndex]

	// Get joke in requested language
	jokeText, exists := randomJoke.Content[lang]
	if !exists {
		jokeText = randomJoke.Content["en"] // Fallback to English
	}

	// Return the joke
	response := JokeResponse{
		Joke:     jokeText,
		Language: lang,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Initialize metrics
	initMetrics() // We don't need to store the exporter, just initialize it

	// Create router
	router := mux.NewRouter()

	// Apply metrics middleware to all routes
	router.Use(metricsMiddleware)

	// API endpoints
	router.HandleFunc("/joke", getRandomJoke).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// Start server
	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
