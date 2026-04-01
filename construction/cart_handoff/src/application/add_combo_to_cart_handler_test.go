package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/infrastructure/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ---------------------------------------------------------------------------
// Fakes / stubs
// ---------------------------------------------------------------------------

// fakeHandoffRepo is an in-memory CartHandoffRecordRepository.
type fakeHandoffRepo struct {
	mu      sync.RWMutex
	records []*domain.CartHandoffRecord
}

func (r *fakeHandoffRepo) Save(_ context.Context, rec *domain.CartHandoffRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
	return nil
}

func (r *fakeHandoffRepo) FindById(_ context.Context, id domain.CartHandoffRecordId) (*domain.CartHandoffRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rec := range r.records {
		if rec.ID() == id {
			return rec, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *fakeHandoffRepo) FindByShopperId(_ context.Context, shopperID domain.ShopperId) ([]*domain.CartHandoffRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.CartHandoffRecord
	for _, rec := range r.records {
		if rec.ShopperID() == shopperID {
			out = append(out, rec)
		}
	}
	return out, nil
}

func (r *fakeHandoffRepo) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.records)
}

func (r *fakeHandoffRepo) last() *domain.CartHandoffRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.records) == 0 {
		return nil
	}
	return r.records[len(r.records)-1]
}

// stubComboPort returns a fixed item list.
type stubComboPort struct {
	items []domain.CartItem
	err   error
}

func (s *stubComboPort) FetchComboItems(_ context.Context, _, _ string) ([]domain.CartItem, error) {
	return s.items, s.err
}

// stubCartPort returns configurable added/skipped results.
type stubCartPort struct {
	result acl.CartSubmissionResult
	err    error
}

func (s *stubCartPort) BulkAddToCart(_ context.Context, _ []domain.CartItem, _ string) (acl.CartSubmissionResult, error) {
	return s.result, s.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeCartItems(n int) []domain.CartItem {
	items := make([]domain.CartItem, n)
	for i := 0; i < n; i++ {
		items[i] = domain.CartItem{SimpleSku: "sku-" + string(rune('a'+i)), Quantity: 1, Size: "M"}
	}
	return items
}

func makeSkipped(skus []domain.CartItem) []domain.SkippedItem {
	out := make([]domain.SkippedItem, len(skus))
	for i, it := range skus {
		out[i] = domain.SkippedItem{SimpleSku: it.SimpleSku, Reason: "out_of_stock"}
	}
	return out
}

func newHandler(repo *fakeHandoffRepo, comboPort acl.ComboPortfolioPort, cartPort acl.PlatformCartPort) *application.AddComboToCartHandler {
	resolutionSvc := services.NewComboResolutionService(comboPort)
	submissionSvc := services.NewCartSubmissionService(cartPort)
	return application.NewAddComboToCartHandler(repo, resolutionSvc, submissionSvc)
}

// ---------------------------------------------------------------------------
// 5.1 Both ComboId + InlineItems -> ErrInvalidHandoffSource, no repo write
// ---------------------------------------------------------------------------

func TestAddComboToCart_BothSourcesRejected(t *testing.T) {
	repo := &fakeHandoffRepo{}
	h := newHandler(repo, &stubComboPort{items: makeCartItems(2)}, &stubCartPort{})
	_, err := h.Handle(context.Background(), application.AddComboToCartCommand{
		ShopperID:   "s1",
		ComboId:     "combo-abc",
		InlineItems: makeCartItems(2),
	})
	assert.ErrorIs(t, err, domain.ErrInvalidHandoffSource)
	assert.Equal(t, 0, repo.count(), "no audit record should be written on invalid source")
}

// ---------------------------------------------------------------------------
// 5.2 Neither source -> ErrInvalidHandoffSource, no repo write
// ---------------------------------------------------------------------------

func TestAddComboToCart_NeitherSourceRejected(t *testing.T) {
	repo := &fakeHandoffRepo{}
	h := newHandler(repo, &stubComboPort{}, &stubCartPort{})
	_, err := h.Handle(context.Background(), application.AddComboToCartCommand{
		ShopperID: "s1",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidHandoffSource)
	assert.Equal(t, 0, repo.count())
}

// ---------------------------------------------------------------------------
// 5.3 All items added -> status ok, audit record persisted
// ---------------------------------------------------------------------------

func TestAddComboToCart_AllAddedStatusOk(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 8).Draw(t, "n")
		items := makeCartItems(n)
		repo := &fakeHandoffRepo{}
		h := newHandler(repo,
			&stubComboPort{},
			&stubCartPort{result: acl.CartSubmissionResult{AddedItems: items, SkippedItems: nil}},
		)
		result, err := h.Handle(context.Background(), application.AddComboToCartCommand{
			ShopperID:   "s1",
			InlineItems: items,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.Status != domain.HandoffStatusOk {
			t.Fatalf("expected ok status, got %v", result.Status)
		}
		if len(result.AddedItems) != n {
			t.Fatalf("expected %d added items, got %d", n, len(result.AddedItems))
		}
		if repo.count() != 1 {
			t.Fatalf("expected 1 audit record, got %d", repo.count())
		}
	})
}

// ---------------------------------------------------------------------------
// 5.4 Some items skipped -> status partial, audit counts consistent
// ---------------------------------------------------------------------------

func TestAddComboToCart_PartialSkipStatusPartial(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nAdded := rapid.IntRange(1, 5).Draw(t, "nAdded")
		nSkipped := rapid.IntRange(1, 5).Draw(t, "nSkipped")
		added := makeCartItems(nAdded)
		skipped := makeSkipped(makeCartItems(nSkipped))
		repo := &fakeHandoffRepo{}
		h := newHandler(repo,
			&stubComboPort{},
			&stubCartPort{result: acl.CartSubmissionResult{AddedItems: added, SkippedItems: skipped}},
		)
		result, err := h.Handle(context.Background(), application.AddComboToCartCommand{
			ShopperID:   "s1",
			InlineItems: added,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != domain.HandoffStatusPartial {
			t.Fatalf("expected partial, got %v", result.Status)
		}
		if len(result.AddedItems) != nAdded {
			t.Fatalf("addedItems count mismatch: got %d, want %d", len(result.AddedItems), nAdded)
		}
		if len(result.SkippedItems) != nSkipped {
			t.Fatalf("skippedItems count mismatch: got %d, want %d", len(result.SkippedItems), nSkipped)
		}
	})
}

// ---------------------------------------------------------------------------
// 5.5 Platform total failure -> status failed, all skipped as platform_error
// ---------------------------------------------------------------------------

func TestAddComboToCart_PlatformFailureStatusFailed(t *testing.T) {
	items := makeCartItems(3)
	repo := &fakeHandoffRepo{}
	h := newHandler(repo,
		&stubComboPort{},
		&stubCartPort{err: errors.New("platform down")},
	)
	_, err := h.Handle(context.Background(), application.AddComboToCartCommand{
		ShopperID:   "s1",
		InlineItems: items,
	})
	// The handler wraps the platform error
	assert.Error(t, err)
	// Audit record MUST still be persisted
	assert.Equal(t, 1, repo.count(), "audit record should be written even on platform failure")
	rec := repo.last()
	require.NotNil(t, rec)
	assert.Equal(t, domain.HandoffStatusFailed, rec.Status())
	for _, sk := range rec.SkippedItems() {
		assert.Equal(t, "platform_error", sk.Reason)
	}
}

// ---------------------------------------------------------------------------
// 5.6 Audit record always persisted (property over all success/failure combos)
// ---------------------------------------------------------------------------

func TestAddComboToCart_AuditRecordAlwaysPersisted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nAdded := rapid.IntRange(0, 5).Draw(t, "nAdded")
		nSkipped := rapid.IntRange(0, 5).Draw(t, "nSkipped")
		added := makeCartItems(nAdded)
		skipped := makeSkipped(makeCartItems(nSkipped))
		repo := &fakeHandoffRepo{}
		h := newHandler(repo,
			&stubComboPort{},
			&stubCartPort{result: acl.CartSubmissionResult{AddedItems: added, SkippedItems: skipped}},
		)
		// We need at least 1 item to have a valid inline_items source
		inline := makeCartItems(max(nAdded+nSkipped, 1))
		h.Handle(context.Background(), application.AddComboToCartCommand{ //nolint:errcheck
			ShopperID:   "s1",
			InlineItems: inline,
		})
		if repo.count() != 1 {
			t.Fatalf("expected exactly 1 audit record, got %d", repo.count())
		}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// 5.7 Status in HandoffResult matches status in persisted CartHandoffRecord
// ---------------------------------------------------------------------------

func TestAddComboToCart_ResultStatusMatchesPersistedRecordStatus(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nAdded := rapid.IntRange(0, 5).Draw(t, "nAdded")
		nSkipped := rapid.IntRange(0, 5).Draw(t, "nSkipped")
		added := makeCartItems(nAdded)
		skipped := makeSkipped(makeCartItems(nSkipped))
		repo := &fakeHandoffRepo{}
		if nAdded == 0 && nSkipped == 0 {
			// skip degenerate case where platform returns nothing added
			// (would also be failed status in both places but not interesting to test here)
			return
		}
		h := newHandler(repo,
			&stubComboPort{},
			&stubCartPort{result: acl.CartSubmissionResult{AddedItems: added, SkippedItems: skipped}},
		)
		inline := makeCartItems(max(nAdded, 1))
		result, err := h.Handle(context.Background(), application.AddComboToCartCommand{
			ShopperID:   "s1",
			InlineItems: inline,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rec := repo.last()
		if rec == nil {
			t.Fatal("expected persisted record")
		}
		if result.Status != rec.Status() {
			t.Fatalf("HandoffResult.Status %q != persisted status %q", result.Status, rec.Status())
		}
	})
}
