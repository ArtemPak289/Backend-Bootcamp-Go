package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"store_api/internal/config"
	"store_api/internal/database"
	"store_api/internal/handlers"
	"store_api/internal/mappers"
	"store_api/internal/models/dto"
	"store_api/internal/repository"
	"store_api/internal/router"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func setupTestServer(t *testing.T) (*mux.Router, func()) {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	addrRepo := repository.NewAddressRepository(db)
	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	imageRepo := repository.NewImageRepository(db)

	clientMapper := mappers.NewClientMapper()
	productMapper := mappers.NewProductMapper()
	supplierMapper := mappers.NewSupplierMapper()
	imageMapper := mappers.NewImageMapper()

	clientH := handlers.NewClientHandler(clientRepo, addrRepo, clientMapper)
	productH := handlers.NewProductHandler(productRepo, supplierRepo, imageRepo, productMapper)
	supplierH := handlers.NewSupplierHandler(supplierRepo, addrRepo, supplierMapper)
	imageH := handlers.NewImageHandler(imageRepo, productRepo, imageMapper)

	r := router.NewRouter(clientH, productH, supplierH, imageH)

	cleanup := func() {
		db.Close()
	}

	return r, cleanup
}

func TestSwaggerEndpoint(t *testing.T) {
	r, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for swagger UI, got %d", rr.Code)
	}

	reqJSON, _ := http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rrJSON := httptest.NewRecorder()
	r.ServeHTTP(rrJSON, reqJSON)

	if rrJSON.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for swagger doc.json, got %d", rrJSON.Code)
	}
}

func TestClientsFlow(t *testing.T) {
	r, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Add Client
	clientReq := dto.ClientRequestDTO{
		ClientName:    "Alice",
		ClientSurname: "Smith",
		Birthday:      "1992-08-24",
		Gender:        "female",
		Address: &dto.AddressDTO{
			Country: "USA",
			City:    "Boston",
			Street:  "Beacon St, 15",
		},
	}
	body, _ := json.Marshal(clientReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 for AddClient, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdClient dto.ClientResponseDTO
	if err := json.Unmarshal(rr.Body.Bytes(), &createdClient); err != nil {
		t.Fatalf("Failed to parse created client: %v", err)
	}
	if createdClient.ClientName != "Alice" || createdClient.ClientSurname != "Smith" {
		t.Fatalf("Unexpected client data: %+v", createdClient)
	}

	// 2. Search client by first and last name
	searchURL := fmt.Sprintf("/api/v1/clients/search?client_name=%s&client_surname=%s", "Alice", "Smith")
	req, _ = http.NewRequest(http.MethodGet, searchURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for search clients, got %d: %s", rr.Code, rr.Body.String())
	}
	var searchResults []dto.ClientResponseDTO
	if err := json.Unmarshal(rr.Body.Bytes(), &searchResults); err != nil {
		t.Fatalf("Failed to parse search results: %v", err)
	}
	if len(searchResults) == 0 {
		t.Fatalf("Expected at least one search result, got 0")
	}

	// 3. Search non-existent client (must return empty list [])
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/clients/search?client_name=NonExistent&client_surname=Nobody", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 with empty list for missing client search, got %d", rr.Code)
	}
	var emptySearch []dto.ClientResponseDTO
	_ = json.Unmarshal(rr.Body.Bytes(), &emptySearch)
	if len(emptySearch) != 0 {
		t.Fatalf("Expected empty list, got %d items", len(emptySearch))
	}

	// 4. Get all clients with pagination
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/clients?limit=5&offset=0", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for get all clients, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Change client address
	newAddr := dto.ClientAddressUpdateDTO{
		Country: "Germany",
		City:    "Munich",
		Street:  "Marienplatz 1",
	}
	addrBody, _ := json.Marshal(newAddr)
	updateURL := fmt.Sprintf("/api/v1/clients/%s/address", createdClient.ID)
	req, _ = http.NewRequest(http.MethodPut, updateURL, bytes.NewReader(addrBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for ChangeClientAddress, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Delete client
	deleteURL := fmt.Sprintf("/api/v1/clients/%s", createdClient.ID)
	req, _ = http.NewRequest(http.MethodDelete, deleteURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for DeleteClient, got %d: %s", rr.Code, rr.Body.String())
	}

	// 7. Delete again should return 404
	req, _ = http.NewRequest(http.MethodDelete, deleteURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 for deleting non-existent client, got %d", rr.Code)
	}
}

func TestSuppliersFlow(t *testing.T) {
	r, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Add Supplier
	supReq := dto.SupplierRequestDTO{
		Name:        "Bosch Appliances",
		PhoneNumber: "+49-89-123456",
		Address: &dto.AddressDTO{
			Country: "Germany",
			City:    "Stuttgart",
			Street:  "Robert-Bosch-Platz 1",
		},
	}
	body, _ := json.Marshal(supReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/suppliers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 for AddSupplier, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdSup dto.SupplierResponseDTO
	_ = json.Unmarshal(rr.Body.Bytes(), &createdSup)

	// 2. Get Supplier by ID
	supURL := fmt.Sprintf("/api/v1/suppliers/%s", createdSup.ID)
	req, _ = http.NewRequest(http.MethodGet, supURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GetSupplierByID, got %d: %s", rr.Code, rr.Body.String())
	}

	// 3. Get all suppliers
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/suppliers", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GetAllSuppliers, got %d", rr.Code)
	}

	// 4. Change supplier address
	newAddr := dto.SupplierAddressUpdateDTO{
		Country: "Germany",
		City:    "Frankfurt",
		Street:  "Mainzer Landstrasse 10",
	}
	addrBody, _ := json.Marshal(newAddr)
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/suppliers/%s/address", createdSup.ID), bytes.NewReader(addrBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for ChangeSupplierAddress, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Delete Supplier
	req, _ = http.NewRequest(http.MethodDelete, supURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for DeleteSupplier, got %d", rr.Code)
	}

	// 6. Delete again -> 404
	req, _ = http.NewRequest(http.MethodDelete, supURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 for DeleteSupplier missing, got %d", rr.Code)
	}
}

func TestProductsAndImagesFlow(t *testing.T) {
	r, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Add Product
	prodReq := dto.ProductRequestDTO{
		Name:           "Washing Machine Series 6",
		Category:       "Washing Machines",
		Price:          649.99,
		AvailableStock: 10,
	}
	body, _ := json.Marshal(prodReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 for AddProduct, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdProd dto.ProductResponseDTO
	_ = json.Unmarshal(rr.Body.Bytes(), &createdProd)

	// 2. Get Product By ID
	prodURL := fmt.Sprintf("/api/v1/products/%s", createdProd.ID)
	req, _ = http.NewRequest(http.MethodGet, prodURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GetProductByID, got %d", rr.Code)
	}

	// 3. Get All Available Products
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GetAvailableProducts, got %d", rr.Code)
	}

	// 4. Decrease Quantity (valid)
	decReq := dto.ProductDecreaseStockDTO{Amount: 3}
	decBody, _ := json.Marshal(decReq)
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/products/%s/decrease", createdProd.ID), bytes.NewReader(decBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for DecreaseQuantity, got %d: %s", rr.Code, rr.Body.String())
	}
	var decProd dto.ProductResponseDTO
	_ = json.Unmarshal(rr.Body.Bytes(), &decProd)
	if decProd.AvailableStock != 7 {
		t.Fatalf("Expected available_stock 7, got %d", decProd.AvailableStock)
	}

	// 5. Decrease Quantity (insufficient stock -> 400)
	excessDec := dto.ProductDecreaseStockDTO{Amount: 100}
	excessBody, _ := json.Marshal(excessDec)
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/products/%s/decrease", createdProd.ID), bytes.NewReader(excessBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for excessive decrease, got %d", rr.Code)
	}

	// 6. Add Image for Product via Multipart Form
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("product_id", createdProd.ID.String())
	fileWriter, _ := mw.CreateFormFile("image", "test.jpg")
	sampleBytes := []byte("FAKE_JPEG_IMAGE_DATA_12345")
	_, _ = fileWriter.Write(sampleBytes)
	_ = mw.Close()

	req, _ = http.NewRequest(http.MethodPost, "/api/v1/images", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 for AddImage multipart, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdImage dto.ImageResponseDTO
	_ = json.Unmarshal(rr.Body.Bytes(), &createdImage)
	if createdImage.ID == uuid.Nil {
		t.Fatalf("Expected valid image ID, got nil")
	}

	// 7. Get Image by Product ID
	prodImgURL := fmt.Sprintf("/api/v1/products/%s/image", createdProd.ID)
	req, _ = http.NewRequest(http.MethodGet, prodImgURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GetImageOfProduct, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("Expected Content-Type application/octet-stream, got %s", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); !bytes.Contains([]byte(cd), []byte("attachment")) {
		t.Fatalf("Expected Content-Disposition to contain attachment, got %s", cd)
	}
	if !bytes.Equal(rr.Body.Bytes(), sampleBytes) {
		t.Fatalf("Image data mismatch: expected %s, got %s", string(sampleBytes), rr.Body.String())
	}

	// 8. Get Image by Image ID
	imgURL := fmt.Sprintf("/api/v1/images/%s", createdImage.ID)
	req, _ = http.NewRequest(http.MethodGet, imgURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GetImageByID, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("Expected Content-Type application/octet-stream, got %s", ct)
	}

	// 9. Replace Image
	newImageStr := "UPDATED_IMAGE_DATA_STRING_67890"
	repReq := dto.ImageReplaceDTO{Image: base64.StdEncoding.EncodeToString([]byte(newImageStr))}
	repBody, _ := json.Marshal(repReq)
	req, _ = http.NewRequest(http.MethodPut, imgURL, bytes.NewReader(repBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for ReplaceImage, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify image content was updated
	req, _ = http.NewRequest(http.MethodGet, imgURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if !bytes.Equal(rr.Body.Bytes(), []byte(newImageStr)) {
		t.Fatalf("Expected replaced image bytes %s, got %s", newImageStr, rr.Body.String())
	}

	// 10. Delete Image
	req, _ = http.NewRequest(http.MethodDelete, imgURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for DeleteImage, got %d", rr.Code)
	}

	// 11. Delete Product
	req, _ = http.NewRequest(http.MethodDelete, prodURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for DeleteProduct, got %d", rr.Code)
	}

	// 12. Delete Product again -> 404
	req, _ = http.NewRequest(http.MethodDelete, prodURL, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404 for DeleteProduct missing, got %d", rr.Code)
	}
}

func TestValidationErrors(t *testing.T) {
	r, cleanup := setupTestServer(t)
	defer cleanup()

	// Client with empty name -> 400
	invalidClient := dto.ClientRequestDTO{
		ClientName: "",
	}
	body, _ := json.Marshal(invalidClient)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for empty client name, got %d", rr.Code)
	}

	// Product with negative price -> 400
	invalidProd := dto.ProductRequestDTO{
		Name:     "Kettle",
		Category: "Appliances",
		Price:    -10.0,
	}
	body, _ = json.Marshal(invalidProd)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for negative price, got %d", rr.Code)
	}

	// Supplier with empty phone -> 400
	invalidSup := dto.SupplierRequestDTO{
		Name:        "Supplier Inc",
		PhoneNumber: "",
	}
	body, _ = json.Marshal(invalidSup)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/suppliers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 for empty phone number, got %d", rr.Code)
	}

	// Non-existent IDs -> 404
	randomUUID := uuid.New().String()
	for _, path := range []string{
		"/api/v1/clients/" + randomUUID,
		"/api/v1/products/" + randomUUID,
		"/api/v1/suppliers/" + randomUUID,
		"/api/v1/images/" + randomUUID,
		"/api/v1/products/" + randomUUID + "/image",
	} {
		req, _ = http.NewRequest(http.MethodGet, path, nil)
		rr = httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("Expected status 404 for %s, got %d", path, rr.Code)
		}
	}
}

func ioReadAll(r io.Reader) []byte {
	b, _ := io.ReadAll(r)
	return b
}
