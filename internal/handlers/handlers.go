package handlers

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"log/slog"
	"net/http"

	"quotemanager/internal/models"
	"quotemanager/internal/repositories"
	"quotemanager/pkg/errors"
)

func AddQuoteHandler(log *slog.Logger, db repositories.DBInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("adding quote handler")
		log.Info("start adding quote")
		var request struct {
			Author string `json:"author"`
			Quote  string `json:"quote"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error("failed to decode request body", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		newQuote := models.Quote{
			Author: request.Author,
			Quote:  request.Quote,
		}

		if err := db.AddQuote(r.Context(), newQuote); err != nil {
			log.Error("failed to add quote", "error", err)
			http.Error(w, "Failed to add quote", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("Quote was added successfully\n"))
		if err != nil {
			log.Error("error writing", "error", err)
		}

		log.Info("Finished adding quote")
	}
}

func GetQuotesHandler(log *slog.Logger, db repositories.DBInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Getting quote data handler")
		log.Info("Started fetching quote")

		filters := models.QuoteFilter{
			Author: r.URL.Query().Get("author"),
		}

		quotes, err := db.GetQuotes(r.Context(), filters)
		if err != nil {
			log.Error("Failed to fetch quotes", "error", err)
			http.Error(w, "Failed to fetch quotes", http.StatusInternalServerError)
			return
		}

		if len(quotes) == 0 {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte("[]\n"))
			if err != nil {
				log.Error("error writing empty response", "error", err)
			}
		} else {
			w.Header().Set("Content-Type", "application/json")

			jsonData, err := json.MarshalIndent(quotes, "", "  ")
			if err != nil {
				log.Error("failed to encode quotes to JSON", "error", err)
				http.Error(w, "Failed to encode quotes", http.StatusInternalServerError)
				return
			}

			_, err = w.Write(jsonData)
			if err != nil {
				log.Error("error writing", "error", err)
			}
		}
		log.Info("Finished fetching quotes")
	}
}

func DeleteQuoteHandler(log *slog.Logger, db repositories.DBInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Deleting quote handler")
		log.Info("Started deleting quote")
		quoteID := r.PathValue("quoteID")

		if err := db.DeleteQuote(r.Context(), quoteID); err != nil {
			if stdErrors.Is(err, errors.ErrQuoteNotFound) {
				log.Warn("The quote to delete is not found", "error", err)
				http.Error(w, "The quote to delete is not found", http.StatusNotFound)
			} else {
				log.Error("failed to delete quote", "error", err)
				http.Error(w, "Failed to delete quote", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		outstr := fmt.Sprintf("quote with id %v was deleted successfully\n", quoteID)
		_, err := w.Write([]byte(outstr))
		if err != nil {
			log.Error("error writing", "error", err)
		}

		log.Info("Finished deleting quote")
	}
}

func GetRandomQuoteHandler(log *slog.Logger, db repositories.DBInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Started getting random quote handler")
		log.Info("Started getting random quote")

		quote, err := db.GetRandomQuote(r.Context())
		if err != nil {
			if stdErrors.Is(err, errors.ErrQuoteNotFound) {
				log.Warn("random quote not found", "error", err)
				http.Error(w, "Random quote not found", http.StatusNotFound)
			} else {
				log.Error("failed to get random quote", "error", err)
				http.Error(w, "Failed to get random quote", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")

		jsonData, err := json.MarshalIndent(quote, "", "  ")
		if err != nil {
			log.Error("failed to encode quotes to JSON", "error", err)
			http.Error(w, "Failed to encode quotes", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(jsonData)
		if err != nil {
			log.Error("error writing", "error", err)
		}
		log.Info("Finished getting random quote")
	}
}
