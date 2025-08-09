package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/api/dto"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/api/middleware"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// TransactionHandler handles HTTP requests for transaction operations
// @Description Handles transaction creation, retrieval, and listing
// @Tags transactions
type TransactionHandler struct {
	transactionService service.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler with the given TransactionService
func NewTransactionHandler(ts service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: ts,
	}
}

// CreateTransaction handles the creation of a new transaction
// @Summary Create a new transaction
// @Description Creates a new transaction with the provided details
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body dto.CreateTransactionRequest true "Transaction details"
// @Success 201 {object} dto.TransactionResponse "Transaction created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request format"
// @Failure 422 {object} dto.ErrorResponse "Validation error"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get validated data from context
	var req dto.CreateTransactionRequest
	if !middleware.GetValidatedData(r, &req) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request data"})
		return
	}

	// Convert DTO to model
	entry := req.ToModel()

	// Create the transaction
	if err := h.transactionService.CreateEntry(ctx, entry); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	// Convert to response DTO
	resp := dto.ToResponse(entry)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// GetTransaction handles retrieving a transaction by ID
// @Summary Get a transaction by ID
// @Description Retrieves a transaction with the specified ID
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} dto.TransactionResponse "Transaction found"
// @Failure 404 {object} dto.ErrorResponse "Transaction not found"
// @Failure 400 {object} dto.ErrorResponse "Invalid ID format"
// @Router /api/v1/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transactionID := chi.URLParam(r, "id")
	if transactionID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Transaction ID is required"})
		return
	}

	// Get transaction from service
	transaction, err := h.transactionService.GetEntryByID(ctx, transactionID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	if transaction == nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Transaction not found"})
		return
	}

	// Convert to response DTO
	resp := dto.ToResponse(transaction)
	render.JSON(w, r, resp)
}

// ListTransactions handles listing transactions with optional filtering
// @Summary List transactions
// @Description Retrieves a paginated list of transactions with optional date filtering
// @Tags transactions
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)" format(date-time)
// @Param end_date query string false "End date (RFC3339 format)" format(date-time)
// @Param page query int false "Page number" minimum(1) default(1)
// @Param page_size query int false "Items per page" minimum(1) maximum(100) default(20)
// @Success 200 {object} dto.TransactionsListResponse "List of transactions"
// @Failure 400 {object} dto.ErrorResponse "Invalid parameters"
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Default to last 30 days if no dates provided
	if startDateStr == "" || endDateStr == "" {
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		startDateStr = startDate.Format(time.RFC3339)
		endDateStr = endDate.Format(time.RFC3339)
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid start_date format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid end_date format. Use RFC3339 format (e.g., 2023-01-31T23:59:59Z)"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 20
	}

	// Get transactions from service
	transactions, total, err := h.transactionService.GetEntriesByDateRange(ctx, startDate, endDate, page, pageSize)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	// Convert to response DTOs
	responses := make([]dto.TransactionResponse, 0, len(transactions))
	for _, t := range transactions {
		resp := dto.ToResponse(t)
		responses = append(responses, *resp)
	}

	// Calculate total pages
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// Prepare response
	resp := dto.TransactionsListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	}

	render.JSON(w, r, resp)
}

// RegisterRoutes registers transaction routes to the router
func (h *TransactionHandler) RegisterRoutes(router chi.Router) {
	router.Route("/api/v1/transactions", func(r chi.Router) {
		// Apply JSON middleware
		r.Use(middleware.JSONMiddleware)
		r.Use(middleware.ErrorHandler)

		// Create transaction with validation
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var req dto.CreateTransactionRequest
			middleware.ValidateRequest(h.CreateTransaction, &req)(w, r)
		})

		// Get transaction by ID
		r.Get("/{id}", h.GetTransaction)

		// List transactions with pagination
		r.Get("/", h.ListTransactions)
	})
}
