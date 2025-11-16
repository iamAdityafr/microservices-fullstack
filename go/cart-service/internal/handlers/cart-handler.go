package handlers

import (
	"bck/auth/authpb"
	"bck/cart/cartpb"
	"bck/product/productpb"
	"cart-service/internal/database"
	"cart-service/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type CartHandler struct {
	cartpb.UnimplementedCartServiceServer
	repo          database.CartRepository
	productClient productpb.ProductServiceClient
	authClient    authpb.AuthServiceClient
}

func NewCartHandler(repo database.CartRepository, productClient productpb.ProductServiceClient, authClient authpb.AuthServiceClient) *CartHandler {
	return &CartHandler{
		repo:          repo,
		productClient: productClient,
		authClient:    authClient,
	}
}
func (h *CartHandler) GetCartHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		fmt.Println("No authorization cookie:", err)
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: cookie.Value})
	if err != nil {
		fmt.Println("Token validation error:", err)
		http.Error(w, "failed to validate token", http.StatusUnauthorized)
		return
	}
	if !authResp.Valid {
		http.Error(w, "failed to validate token", http.StatusUnauthorized)
		return
	}

	items, err := h.repo.GetCart(r.Context(), authResp.UserId)
	if err != nil {
		http.Error(w, "could not get cart items", http.StatusInternalServerError)
		return
	}

	// fmt.Println("got items:", items)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		fmt.Println("Failed to encode JSON:", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}

}

func (h *CartHandler) AddToCartHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		log.Println("auth validation error:", err)
		http.Error(w, "failed to validate token", http.StatusUnauthorized)
		return
	}
	if !authResp.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userID := authResp.UserId
	type AddToCartRequest struct {
		ProductID string `json:"product_id"`
	}

	var req AddToCartRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	productID := req.ProductID
	log.Println("Fetching product with ID:", productID)

	productResp, err := h.productClient.GetProductById(ctx, &productpb.GetProductByIdRequest{
		Id: productID,
	})
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	item := &models.CartItem{
		ProductID:  productID,
		Name:       productResp.Product.Name,
		PriceCents: productResp.Product.Pricecents,
		Image:      productResp.Product.Image,
		Quantity:   1,
		AddedAt:    time.Now(),
	}

	addedItem, err := h.repo.AddToCart(context.Background(), userID, item)
	if err != nil {
		http.Error(w, "fail add item to cart", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addedItem)
	w.WriteHeader(http.StatusOK)
}
func (h *CartHandler) RemoveFromCartHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
	}
	token := cookie.Value

	authResp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		http.Error(w, "fail to validate token", http.StatusUnauthorized)
		return
	}
	if !authResp.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userID := authResp.UserId

	type RemoveFromCartRequest struct {
		ProductID string `json:"product_id"`
	}
	var req RemoveFromCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.repo.RemoveFromCart(context.Background(), userID, req.ProductID)
	if err != nil {
		http.Error(w, "failed to remove item from cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Item removed",
	})
}

// GRPC HANDLERS
func (h *CartHandler) AddToCart(ctx context.Context, req *cartpb.AddToCartRequest) (*cartpb.AddToCartResponse, error) {
	item := &models.CartItem{
		ProductID:  req.Item.ProductId,
		Name:       req.Item.Name,
		Image:      req.Item.Image,
		PriceCents: req.Item.PriceCents,
		Quantity:   req.Item.Quantity,
		AddedAt:    time.Now(),
	}

	addedItem, err := h.repo.AddToCart(ctx, req.UserId, item)
	if err != nil {
		log.Printf("Error adding to cart: %v", err)
		return nil, err
	}

	resp := &cartpb.AddToCartResponse{
		Item: &cartpb.CartItem{
			ProductId:  addedItem.ProductID,
			Name:       addedItem.Name,
			Image:      addedItem.Image,
			PriceCents: addedItem.PriceCents,
			Quantity:   addedItem.Quantity,
		},
	}

	return resp, nil
}

func (h *CartHandler) RemoveFromCart(ctx context.Context, req *cartpb.RemoveFromCartRequest) (*cartpb.RemoveFromCartResponse, error) {
	err := h.repo.RemoveFromCart(ctx, req.UserId, req.ProductId)
	if err != nil {
		log.Printf("Error remove from cart: %v", err)
		return &cartpb.RemoveFromCartResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &cartpb.RemoveFromCartResponse{
		Success: true,
		Message: "Item removed successfully",
	}, nil
}

func (h *CartHandler) GetCart(ctx context.Context, req *cartpb.GetCartRequest) (*cartpb.GetCartResponse, error) {
	cart, err := h.repo.GetCart(ctx, req.UserId)
	if err != nil {
		log.Printf("could not fetch cart: %v", err)
		return nil, err
	}

	items := make([]*cartpb.CartItem, 0, len(cart.Items))
	for _, i := range cart.Items {
		items = append(items, &cartpb.CartItem{
			ProductId:  i.ProductID,
			Name:       i.Name,
			Image:      i.Image,
			PriceCents: i.PriceCents,
			Quantity:   i.Quantity,
		})
	}

	resp := &cartpb.GetCartResponse{
		Cart: &cartpb.Cart{
			Id:     cart.ID.Hex(),
			UserId: cart.UserID,
			Items:  items,
		},
	}

	return resp, nil
}
