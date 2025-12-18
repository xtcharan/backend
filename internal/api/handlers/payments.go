package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/college-event-backend/internal/models"
	"github.com/yourusername/college-event-backend/pkg/database"
)

type PaymentHandler struct {
	db        *database.DB
	keyID     string
	keySecret string
}

func NewPaymentHandler(db *database.DB) *PaymentHandler {
	return &PaymentHandler{
		db:        db,
		keyID:     os.Getenv("RAZORPAY_KEY_ID"),
		keySecret: os.Getenv("RAZORPAY_KEY_SECRET"),
	}
}

// CreateOrder creates a Razorpay order for event payment
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("unauthorized"),
		})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid request body"),
		})
		return
	}

	// Get event details
	var event models.Event
	err := h.db.QueryRow(`
		SELECT id, title, is_paid_event, event_amount, currency
		FROM events
		WHERE id = $1 AND deleted_at IS NULL
	`, req.EventID).Scan(&event.ID, &event.Title, &event.IsPaidEvent, &event.EventAmount, &event.Currency)

	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   strPtr("event not found"),
		})
		return
	}

	if !event.IsPaidEvent {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("this event does not require payment"),
		})
		return
	}

	// Check if user already has a successful payment
	var existingPayment string
	err = h.db.QueryRow(`
		SELECT status FROM event_payments
		WHERE event_id = $1 AND user_id = $2 AND status = 'paid'
	`, req.EventID, userID).Scan(&existingPayment)

	if err == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("you have already paid for this event"),
		})
		return
	}

	// Convert amount to paise (Razorpay expects amount in smallest currency unit)
	amountInPaise := int(*event.EventAmount * 100)
	currency := "INR"
	if event.Currency != nil {
		currency = *event.Currency
	}

	// Create order via Razorpay API
	orderID, err := h.createRazorpayOrder(amountInPaise, currency)
	if err != nil {
		fmt.Printf("Razorpay order creation error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to create payment order"),
		})
		return
	}

	// Store pending payment record
	paymentID := uuid.New()
	_, err = h.db.Exec(`
		INSERT INTO event_payments (id, event_id, user_id, razorpay_order_id, amount, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
	`, paymentID, req.EventID, userID, orderID, *event.EventAmount, currency)

	if err != nil {
		fmt.Printf("Failed to store payment record: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to create payment record"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.CreateOrderResponse{
			OrderID:  orderID,
			Amount:   amountInPaise,
			Currency: currency,
			KeyID:    h.keyID,
			EventID:  req.EventID.String(),
		},
	})
}

// VerifyPayment verifies the payment signature and registers user for event
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("unauthorized"),
		})
		return
	}

	var req models.VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid request body"),
		})
		return
	}

	// Verify signature
	data := req.RazorpayOrderID + "|" + req.RazorpayPaymentID
	h256 := hmac.New(sha256.New, []byte(h.keySecret))
	h256.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h256.Sum(nil))

	if expectedSignature != req.RazorpaySignature {
		// Update payment status to failed
		h.db.Exec(`
			UPDATE event_payments
			SET status = 'failed', failure_reason = 'Invalid signature', updated_at = CURRENT_TIMESTAMP
			WHERE razorpay_order_id = $1
		`, req.RazorpayOrderID)

		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("payment verification failed: invalid signature"),
		})
		return
	}

	// Parse event ID
	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid event ID"),
		})
		return
	}

	// Update payment record
	_, err = h.db.Exec(`
		UPDATE event_payments
		SET razorpay_payment_id = $1, razorpay_signature = $2, status = 'paid', updated_at = CURRENT_TIMESTAMP
		WHERE razorpay_order_id = $3 AND user_id = $4
	`, req.RazorpayPaymentID, req.RazorpaySignature, req.RazorpayOrderID, userID)

	if err != nil {
		fmt.Printf("Failed to update payment record: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   strPtr("failed to update payment record"),
		})
		return
	}

	// Register user for event
	_, err = h.db.Exec(`
		INSERT INTO event_registrations (event_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (event_id, user_id) DO NOTHING
	`, eventID, userID)

	if err != nil {
		fmt.Printf("Failed to register user for event: %v\n", err)
	}

	// Update event participant count
	h.db.Exec(`
		UPDATE events 
		SET current_participants = current_participants + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, eventID)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "payment verified and registration complete",
	})
}

// GetPaymentStatus checks if user has paid for an event
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   strPtr("unauthorized"),
		})
		return
	}

	eventIDStr := c.Param("event_id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   strPtr("invalid event ID"),
		})
		return
	}

	var payment models.EventPayment
	err = h.db.QueryRow(`
		SELECT razorpay_payment_id, status
		FROM event_payments
		WHERE event_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, eventID, userID).Scan(&payment.RazorpayPaymentID, &payment.Status)

	if err != nil {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data: models.PaymentStatusResponse{
				HasPaid: false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PaymentStatusResponse{
			HasPaid:   payment.Status == "paid",
			PaymentID: payment.RazorpayPaymentID,
			Status:    &payment.Status,
		},
	})
}

// createRazorpayOrder calls Razorpay API to create an order
func (h *PaymentHandler) createRazorpayOrder(amount int, currency string) (string, error) {
	url := "https://api.razorpay.com/v1/orders"

	payload := map[string]interface{}{
		"amount":   amount,
		"currency": currency,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(h.keyID, h.keySecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("razorpay API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	orderID, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response from Razorpay")
	}

	return orderID, nil
}
