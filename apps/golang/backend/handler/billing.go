package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

type BillingHandler struct {
	billing *usecase.BillingService
}

func NewBillingHandler(billing *usecase.BillingService) *BillingHandler {
	return &BillingHandler{billing: billing}
}

func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateBillingCheckoutSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	url, err := h.billing.CreateCheckoutSession(r.Context(), req.PriceId, ptrValue(req.SuccessUrl), ptrValue(req.CancelUrl))
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrBillingNotConfigured):
			writeError(w, http.StatusInternalServerError, "billing is not configured")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, openapi.CreateBillingCheckoutSessionResponse{Url: url})
}

func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateBillingPortalSessionRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
	}

	url, err := h.billing.CreatePortalSession(r.Context(), ptrValue(req.ReturnUrl))
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrBillingNotConfigured):
			writeError(w, http.StatusInternalServerError, "billing is not configured")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, openapi.CreateBillingPortalSessionResponse{Url: url})
}

func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	sub, plan, err := h.billing.GetSubscription(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := openapi.BillingSubscriptionResponse{
		TenantId:             sub.TenantID,
		StripeCustomerId:     optionalString(sub.StripeCustomerID),
		StripeSubscriptionId: optionalString(sub.StripeSubscriptionID),
		StripePriceId:        optionalString(sub.StripePriceID),
		Status:               sub.SubscriptionStatus,
	}
	if sub.CurrentPeriodEnd != nil {
		resp.CurrentPeriodEnd = sub.CurrentPeriodEnd
	}
	if plan != nil {
		p := toOpenAPIPlan(plan)
		resp.Plan = &p
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sig := r.Header.Get("Stripe-Signature")
	if sig == "" {
		writeError(w, http.StatusBadRequest, "missing Stripe-Signature header")
		return
	}

	if err := h.billing.HandleWebhook(r.Context(), payload, sig); err != nil {
		switch {
		case errors.Is(err, usecase.ErrBillingNotConfigured):
			writeError(w, http.StatusInternalServerError, "billing is not configured")
		default:
			writeError(w, http.StatusBadRequest, "failed to process webhook")
		}
		return
	}
	writeJSON(w, http.StatusOK, openapi.BillingWebhookResponse{Received: true})
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func ptrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
