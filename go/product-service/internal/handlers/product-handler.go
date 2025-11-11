package handlers

import (
	"bck/auth/authpb"
	"bck/product/productpb"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"product-service/internal/database"
	"product-service/internal/kafka"
	"product-service/internal/models"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	productpb.UnimplementedProductServiceServer
	productRepo     database.ProductRepository
	productProducer *kafka.ProductProducer
	authClient      authpb.AuthServiceClient
}

func NewProductHandler(repo database.ProductRepository, authClient authpb.AuthServiceClient, producer *kafka.ProductProducer) *ProductHandler {
	return &ProductHandler{
		productRepo:     repo,
		authClient:      authClient,
		productProducer: producer,
	}
}
func (h *ProductHandler) GetAllProductsHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	products, err := h.productRepo.GetAllProducts(ctx)
	if err != nil {
		http.Error(w, "couldn't fetch products ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) CreateProductHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "missing auth cookie", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: token})
	if err != nil {
		log.Println("auth validation error:", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !authResp.Valid {
		log.Println("invalid token for user:", authResp.UserId)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "cannot parse form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	category := r.FormValue("category")
	priceStr := r.FormValue("pricecents")
	description := r.FormValue("description")

	if name == "" || priceStr == "" || category == "" || description == "" {
		http.Error(w, "uhh! some values are missing", http.StatusBadRequest)
		return
	}

	priceCents, err := strconv.ParseInt(priceStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid pricecents value", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads" // not using any cloud service yet...

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			http.Error(w, "cannot create upload dir", http.StatusInternalServerError)
			return
		}
	}

	imagePath := filepath.Join(uploadDir, handler.Filename)
	out, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "cannot save image", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "cannot save image", http.StatusInternalServerError)
		return
	}
	p := &models.Product{
		Name:        name,
		Category:    category,
		Image:       imagePath,
		PriceCents:  priceCents,
		Description: description,
	}

	createdProduct, err := h.productRepo.CreateProduct(r.Context(), p)
	if err != nil {
		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)
}

func (h *ProductHandler) GetProductByIdHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "missing auth cookie", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: token})
	if err != nil || !authResp.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	product, err := h.productRepo.GetProductById(r.Context(), id)
	if err != nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) SearchProductHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "missing auth cookie", http.StatusUnauthorized)
		return
	}

	token := cookie.Value
	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: token})
	if err != nil || !authResp.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing search query", http.StatusBadRequest)
		return
	}

	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"category": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	products, err := h.productRepo.SearchProduct(r.Context(), filter, 0, 0)
	if err != nil {
		http.Error(w, "couldn't search for products", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) UpdateProductHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	token := cookie.Value
	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: token})
	if err != nil || !authResp.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing product ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name        *string `json:"name,omitempty"`
		Category    *string `json:"category,omitempty"`
		Image       *string `json:"image,omitempty"`
		PriceCents  *int64  `json:"pricecents,omitempty"`
		Description *string `json:"description,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	changes := make(map[string]any)
	if req.Name != nil {
		changes["name"] = *req.Name
	}
	if req.Category != nil {
		changes["category"] = *req.Category
	}
	if req.Image != nil {
		changes["image"] = *req.Image
	}
	if req.PriceCents != nil {
		changes["pricecents"] = *req.PriceCents
	}
	if req.Description != nil {
		changes["description"] = *req.Description
	}
	if len(changes) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	updatedProduct, err := h.productRepo.UpdateProduct(r.Context(), id, changes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event := models.ProductUpdatedEvent{
		ID:         updatedProduct.ID.Hex(),
		Name:       updatedProduct.Name,
		Category:   updatedProduct.Category,
		Image:      updatedProduct.Image,
		PriceCents: updatedProduct.PriceCents,
		UpdatedAt:  updatedProduct.UpdatedAt,
	}

	if err := h.productProducer.PublishProductUpdated(r.Context(), event); err != nil {
		log.Printf("Failed to publish ProductUpdatedEvent for product %s: %v", updatedProduct.ID.Hex(), err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProduct)
}

func (h *ProductHandler) DeleteProductHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "missing auth cookie", http.StatusUnauthorized)
		return
	}
	token := cookie.Value
	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: token})
	if err != nil || !authResp.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id missing", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeleteProduct(r.Context(), id); err != nil {
		http.Error(w, "product not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("product deleted successfully")
}

// GRPC Handlers
func (h *ProductHandler) GetProductById(ctx context.Context, req *productpb.GetProductByIdRequest) (*productpb.GetProductByIdResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing product ID")
	}
	product, err := h.productRepo.GetProductById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &productpb.GetProductByIdResponse{
		Product: &productpb.Product{
			Id:          product.ID.Hex(),
			Name:        product.Name,
			Category:    product.Category,
			Image:       product.Image,
			Pricecents:  product.PriceCents,
			Description: product.Description,
			CreatedAt:   product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
	changes := map[string]interface{}{}
	if req.Name != "" {
		changes["name"] = req.Name
	}
	if req.Category != "" {
		changes["category"] = req.Category
	}
	if req.Image != "" {
		changes["image"] = req.Image
	}
	if req.Pricecents != 0 {
		changes["pricecents"] = req.Pricecents
	}
	if req.Description != "" {
		changes["description"] = req.Description
	}
	changes["updated_at"] = time.Now()

	updatedProduct, err := h.productRepo.UpdateProduct(ctx, req.Id, changes)
	if err != nil {
		return nil, err
	}

	event := models.ProductUpdatedEvent{
		ID:         updatedProduct.ID.Hex(),
		Name:       updatedProduct.Name,
		Category:   updatedProduct.Category,
		Image:      updatedProduct.Image,
		PriceCents: updatedProduct.PriceCents,
		UpdatedAt:  updatedProduct.UpdatedAt,
	}

	if err := h.productProducer.PublishProductUpdated(ctx, event); err != nil {
		log.Printf("Failed to publish ProductUpdatedEvent for product %s: %v", updatedProduct.ID.Hex(), err)
	}

	return &productpb.UpdateProductResponse{
		Product: &productpb.Product{
			Id:          updatedProduct.ID.Hex(),
			Name:        updatedProduct.Name,
			Category:    updatedProduct.Category,
			Image:       updatedProduct.Image,
			Pricecents:  updatedProduct.PriceCents,
			Description: updatedProduct.Description,
			CreatedAt:   updatedProduct.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   updatedProduct.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}
