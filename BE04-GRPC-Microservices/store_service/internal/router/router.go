package router

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "store_api/docs" // Swagger generated docs
	"store_api/internal/auth"
	"store_api/internal/handlers"
)

func NewRouter(
	clientH *handlers.ClientHandler,
	productH *handlers.ProductHandler,
	supplierH *handlers.SupplierHandler,
	imageH *handlers.ImageHandler,
	authH *handlers.AuthHandler,
	authMiddleware *auth.AuthMiddleware,
) *mux.Router {
	r := mux.NewRouter()

	// Global Middlewares
	r.Use(loggingMiddleware)
	r.Use(recoveryMiddleware)
	r.Use(authMiddleware.Middleware)

	// Health check endpoint
	r.HandleFunc("/health", auth.WithAuth(auth.Public(), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})).Methods(http.MethodGet)

	// Routing aliases
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	// Swagger documentation
	r.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Public Authentication Endpoints on root paths
	r.HandleFunc("/register", auth.WithAuth(auth.Public(), authH.Register)).Methods(http.MethodPost)
	r.HandleFunc("/auth", auth.WithAuth(auth.Public(), authH.Login)).Methods(http.MethodPost)
	r.HandleFunc("/reset", auth.WithAuth(auth.Public(), authH.ResetPassword)).Methods(http.MethodPost)
	r.HandleFunc("/change-password", auth.WithAuth(auth.Public(), authH.ChangePassword)).Methods(http.MethodPost)

	// API v1 Subrouter
	api := r.PathPrefix("/api/v1").Subrouter()

	// Public Authentication Endpoints on /api/v1 prefix
	api.HandleFunc("/register", auth.WithAuth(auth.Public(), authH.Register)).Methods(http.MethodPost)
	api.HandleFunc("/auth", auth.WithAuth(auth.Public(), authH.Login)).Methods(http.MethodPost)
	api.HandleFunc("/reset", auth.WithAuth(auth.Public(), authH.ResetPassword)).Methods(http.MethodPost)
	api.HandleFunc("/change-password", auth.WithAuth(auth.Public(), authH.ChangePassword)).Methods(http.MethodPost)

	// Protected Clients endpoints (Self-written authorization attribute: Authorize())
	api.HandleFunc("/clients/search", auth.WithAuth(auth.Authorize(), clientH.GetClientsByName)).Methods(http.MethodGet)
	api.HandleFunc("/clients", auth.WithAuth(auth.Authorize(), clientH.AddClient)).Methods(http.MethodPost)
	api.HandleFunc("/clients", auth.WithAuth(auth.Authorize(), clientH.GetClients)).Methods(http.MethodGet)
	api.HandleFunc("/clients/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), clientH.GetClientByID)).Methods(http.MethodGet)
	api.HandleFunc("/clients/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), clientH.DeleteClient)).Methods(http.MethodDelete)
	api.HandleFunc("/clients/{id:[0-9a-fA-F-]+}/address", auth.WithAuth(auth.Authorize(), clientH.ChangeClientAddress)).Methods(http.MethodPut, http.MethodPatch)

	// Protected Products endpoints
	api.HandleFunc("/products/available", auth.WithAuth(auth.Authorize(), productH.GetAvailableProducts)).Methods(http.MethodGet)
	api.HandleFunc("/products", auth.WithAuth(auth.Authorize(), productH.AddProduct)).Methods(http.MethodPost)
	api.HandleFunc("/products", auth.WithAuth(auth.Authorize(), productH.GetAvailableProducts)).Methods(http.MethodGet)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}/decrease", auth.WithAuth(auth.Authorize(), productH.DecreaseQuantity)).Methods(http.MethodPatch, http.MethodPost)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), productH.DecreaseQuantity)).Methods(http.MethodPatch)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), productH.GetProductByID)).Methods(http.MethodGet)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), productH.DeleteProduct)).Methods(http.MethodDelete)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}/image", auth.WithAuth(auth.Authorize(), imageH.GetImageOfProduct)).Methods(http.MethodGet)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}/images", auth.WithAuth(auth.Authorize(), imageH.AddImage)).Methods(http.MethodPost)

	// Protected Suppliers endpoints
	api.HandleFunc("/suppliers", auth.WithAuth(auth.Authorize(), supplierH.AddSupplier)).Methods(http.MethodPost)
	api.HandleFunc("/suppliers", auth.WithAuth(auth.Authorize(), supplierH.GetAllSuppliers)).Methods(http.MethodGet)
	api.HandleFunc("/suppliers/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), supplierH.GetSupplierByID)).Methods(http.MethodGet)
	api.HandleFunc("/suppliers/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), supplierH.DeleteSupplier)).Methods(http.MethodDelete)
	api.HandleFunc("/suppliers/{id:[0-9a-fA-F-]+}/address", auth.WithAuth(auth.Authorize(), supplierH.ChangeSupplierAddress)).Methods(http.MethodPut, http.MethodPatch)

	// Protected Images endpoints
	api.HandleFunc("/images/product/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), imageH.GetImageOfProduct)).Methods(http.MethodGet)
	api.HandleFunc("/images", auth.WithAuth(auth.Authorize(), imageH.AddImage)).Methods(http.MethodPost)
	api.HandleFunc("/images/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), imageH.GetImageByID)).Methods(http.MethodGet)
	api.HandleFunc("/images/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), imageH.ReplaceImage)).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/images/{id:[0-9a-fA-F-]+}", auth.WithAuth(auth.Authorize(), imageH.DeleteImage)).Methods(http.MethodDelete)

	return r
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "" {
			r.RequestURI = r.URL.RequestURI()
		}
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Panic recovered: %v\nStack: %s", rec, debug.Stack())
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
