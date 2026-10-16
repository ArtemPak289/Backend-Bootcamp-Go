package router

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "store_api/docs" // Swagger generated docs
	"store_api/internal/handlers"
)

func NewRouter(
	clientH *handlers.ClientHandler,
	productH *handlers.ProductHandler,
	supplierH *handlers.SupplierHandler,
	imageH *handlers.ImageHandler,
) *mux.Router {
	r := mux.NewRouter()

	// Global Middlewares
	r.Use(loggingMiddleware)
	r.Use(recoveryMiddleware)

	// Swagger documentation
	r.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API v1 Subrouter
	api := r.PathPrefix("/api/v1").Subrouter()

	// Clients endpoints
	api.HandleFunc("/clients/search", clientH.GetClientsByName).Methods(http.MethodGet)
	api.HandleFunc("/clients", clientH.AddClient).Methods(http.MethodPost)
	api.HandleFunc("/clients", clientH.GetClients).Methods(http.MethodGet)
	api.HandleFunc("/clients/{id:[0-9a-fA-F-]+}", clientH.GetClientByID).Methods(http.MethodGet)
	api.HandleFunc("/clients/{id:[0-9a-fA-F-]+}", clientH.DeleteClient).Methods(http.MethodDelete)
	api.HandleFunc("/clients/{id:[0-9a-fA-F-]+}/address", clientH.ChangeClientAddress).Methods(http.MethodPut, http.MethodPatch)

	// Products endpoints
	api.HandleFunc("/products/available", productH.GetAvailableProducts).Methods(http.MethodGet)
	api.HandleFunc("/products", productH.AddProduct).Methods(http.MethodPost)
	api.HandleFunc("/products", productH.GetAvailableProducts).Methods(http.MethodGet)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}/decrease", productH.DecreaseQuantity).Methods(http.MethodPatch, http.MethodPost)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}", productH.DecreaseQuantity).Methods(http.MethodPatch)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}", productH.GetProductByID).Methods(http.MethodGet)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}", productH.DeleteProduct).Methods(http.MethodDelete)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}/image", imageH.GetImageOfProduct).Methods(http.MethodGet)
	api.HandleFunc("/products/{id:[0-9a-fA-F-]+}/images", imageH.AddImage).Methods(http.MethodPost)

	// Suppliers endpoints
	api.HandleFunc("/suppliers", supplierH.AddSupplier).Methods(http.MethodPost)
	api.HandleFunc("/suppliers", supplierH.GetAllSuppliers).Methods(http.MethodGet)
	api.HandleFunc("/suppliers/{id:[0-9a-fA-F-]+}", supplierH.GetSupplierByID).Methods(http.MethodGet)
	api.HandleFunc("/suppliers/{id:[0-9a-fA-F-]+}", supplierH.DeleteSupplier).Methods(http.MethodDelete)
	api.HandleFunc("/suppliers/{id:[0-9a-fA-F-]+}/address", supplierH.ChangeSupplierAddress).Methods(http.MethodPut, http.MethodPatch)

	// Images endpoints
	api.HandleFunc("/images/product/{id:[0-9a-fA-F-]+}", imageH.GetImageOfProduct).Methods(http.MethodGet)
	api.HandleFunc("/images", imageH.AddImage).Methods(http.MethodPost)
	api.HandleFunc("/images/{id:[0-9a-fA-F-]+}", imageH.GetImageByID).Methods(http.MethodGet)
	api.HandleFunc("/images/{id:[0-9a-fA-F-]+}", imageH.ReplaceImage).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/images/{id:[0-9a-fA-F-]+}", imageH.DeleteImage).Methods(http.MethodDelete)

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
