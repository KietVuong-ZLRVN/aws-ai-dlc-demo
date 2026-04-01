package application_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/application"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ---------------------------------------------------------------------------
// Fake in-memory repository (test double)
// ---------------------------------------------------------------------------

type fakeComboRepo struct {
	mu     sync.RWMutex
	combos map[string]*domain.Combo
}

func newFakeRepo() *fakeComboRepo {
	return &fakeComboRepo{combos: make(map[string]*domain.Combo)}
}

func (r *fakeComboRepo) Save(_ context.Context, combo *domain.Combo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.combos[combo.ID().String()] = combo
	return nil
}

func (r *fakeComboRepo) FindById(_ context.Context, id domain.ComboId) (*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.combos[id.String()]; ok {
		return c, nil
	}
	return nil, domain.ErrComboNotFound
}

func (r *fakeComboRepo) FindByShopperId(_ context.Context, shopperID domain.ShopperId) ([]*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Combo
	for _, c := range r.combos {
		if c.OwnedBy(shopperID) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeComboRepo) FindByShareToken(_ context.Context, token domain.ShareToken) (*domain.Combo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.combos {
		if c.ShareToken() != nil && *c.ShareToken() == token {
			return c, nil
		}
	}
	return nil, domain.ErrComboNotFound
}

func (r *fakeComboRepo) Delete(_ context.Context, id domain.ComboId) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.combos, id.String())
	return nil
}

func (r *fakeComboRepo) size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.combos)
}

// ---------------------------------------------------------------------------
// Stub product catalog (returns nil — triggers fallback-to-snapshot path)
// ---------------------------------------------------------------------------

type stubCatalogPort struct{}

func (s *stubCatalogPort) FetchProduct(_ context.Context, _ string) (*acl.CatalogProduct, error) {
	return nil, nil
}

func newStubEnrichment() *acl.ComboEnrichmentService {
	return acl.NewComboEnrichmentService(&stubCatalogPort{})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeItems(n int) []domain.ComboItem {
	items := make([]domain.ComboItem, n)
	for i := 0; i < n; i++ {
		items[i] = domain.ComboItem{
			ConfigSku: "cfg",
			SimpleSku: strings.Repeat("x", i+1),
			Name:      "product",
			ImageUrl:  "https://example.com/img.jpg",
			Price:     9.99,
		}
	}
	return items
}

func seedCombo(t interface {
	Helper()
	Fatalf(string, ...interface{})
}, repo *fakeComboRepo, shopperID domain.ShopperId, n int) domain.ComboId {
	t.Helper()
	h := application.NewSaveComboHandler(repo)
	id, err := h.Handle(context.Background(), application.SaveComboCommand{
		ShopperID:  shopperID,
		Name:       "Seed Combo",
		Items:      makeItems(n),
		Visibility: domain.VisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("seedCombo: %v", err)
	}
	return id
}

// ---------------------------------------------------------------------------
// 4.1—4.3 SaveComboHandler
// ---------------------------------------------------------------------------

func TestSaveComboHandler_ValidInputPersistsCombo(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := newFakeRepo()
		h := application.NewSaveComboHandler(repo)
		n := rapid.IntRange(2, 10).Draw(t, "n")
		id, err := h.Handle(context.Background(), application.SaveComboCommand{
			ShopperID:  "shopper1",
			Name:       "My Combo",
			Items:      makeItems(n),
			Visibility: domain.VisibilityPrivate,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if id == "" {
			t.Fatal("expected non-empty ComboId")
		}
		_, ferr := repo.FindById(context.Background(), id)
		if ferr != nil {
			t.Fatalf("expected combo in repo, got %v", ferr)
		}
	})
}

func TestSaveComboHandler_InvalidNameRejectsWithoutPersisting(t *testing.T) {
	repo := newFakeRepo()
	h := application.NewSaveComboHandler(repo)
	_, err := h.Handle(context.Background(), application.SaveComboCommand{
		ShopperID:  "shopper1",
		Name:       "",
		Items:      makeItems(3),
		Visibility: domain.VisibilityPrivate,
	})
	assert.ErrorIs(t, err, domain.ErrInvalidComboName)
	assert.Equal(t, 0, repo.size())
}

func TestSaveComboHandler_TooManyItemsRejectsWithoutPersisting(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := newFakeRepo()
		h := application.NewSaveComboHandler(repo)
		n := rapid.IntRange(11, 30).Draw(t, "n")
		_, err := h.Handle(context.Background(), application.SaveComboCommand{
			ShopperID:  "shopper1",
			Name:       "x",
			Items:      makeItems(n),
			Visibility: domain.VisibilityPrivate,
		})
		if err != domain.ErrInvalidItemCount {
			t.Fatalf("expected ErrInvalidItemCount for %d items, got %v", n, err)
		}
		if repo.size() != 0 {
			t.Fatal("expected empty repo after rejected save")
		}
	})
}

// ---------------------------------------------------------------------------
// 4.4—4.5 RenameComboHandler
// ---------------------------------------------------------------------------

func TestRenameComboHandler_ValidOwnerUpdatesName(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-1", 3)
	h := application.NewRenameComboHandler(repo)
	err := h.Handle(context.Background(), application.RenameComboCommand{
		ComboID:   id,
		ShopperID: "owner-1",
		NewName:   "Renamed",
	})
	require.NoError(t, err)
	combo, _ := repo.FindById(context.Background(), id)
	assert.Equal(t, domain.ComboName("Renamed"), combo.Name())
}

func TestRenameComboHandler_WrongOwnerDenied(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-a", 3)
	h := application.NewRenameComboHandler(repo)
	err := h.Handle(context.Background(), application.RenameComboCommand{
		ComboID:   id,
		ShopperID: "intruder-b",
		NewName:   "Hacked",
	})
	assert.ErrorIs(t, err, domain.ErrComboAccessDenied)
	combo, _ := repo.FindById(context.Background(), id)
	assert.Equal(t, domain.ComboName("Seed Combo"), combo.Name())
}

func TestRenameComboHandler_WrongOwnerDeniedForAnyShopperId(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := newFakeRepo()
		id := seedCombo(t, repo, "real-owner", 3)
		h := application.NewRenameComboHandler(repo)
		fakeOwner := domain.ShopperId("fake-" + rapid.StringOfN(rapid.Rune(), 1, 10, -1).Draw(t, "s"))
		if fakeOwner == "real-owner" {
			return
		}
		err := h.Handle(context.Background(), application.RenameComboCommand{
			ComboID:   id,
			ShopperID: fakeOwner,
			NewName:   "Hacked",
		})
		if err != domain.ErrComboAccessDenied {
			t.Fatalf("expected ErrComboAccessDenied for non-owner %q, got %v", fakeOwner, err)
		}
	})
}

// ---------------------------------------------------------------------------
// 4.6—4.7 DeleteComboHandler
// ---------------------------------------------------------------------------

func TestDeleteComboHandler_OwnerDeletesSuccessfully(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-1", 2)
	h := application.NewDeleteComboHandler(repo)
	err := h.Handle(context.Background(), application.DeleteComboCommand{
		ComboID:   id,
		ShopperID: "owner-1",
	})
	require.NoError(t, err)
	_, err = repo.FindById(context.Background(), id)
	assert.ErrorIs(t, err, domain.ErrComboNotFound)
}

func TestDeleteComboHandler_WrongOwnerDenied(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-1", 2)
	h := application.NewDeleteComboHandler(repo)
	err := h.Handle(context.Background(), application.DeleteComboCommand{
		ComboID:   id,
		ShopperID: "intruder",
	})
	assert.ErrorIs(t, err, domain.ErrComboAccessDenied)
	_, err = repo.FindById(context.Background(), id)
	assert.NoError(t, err, "combo should still exist after denied delete")
}

// ---------------------------------------------------------------------------
// 4.8 ShareComboHandler
// ---------------------------------------------------------------------------

func TestShareComboHandler_OwnerGetsTokenAndComboBecomesPublic(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-1", 3)
	h := application.NewShareComboHandler(repo, services.NewShareTokenService(repo))
	token, err := h.Handle(context.Background(), application.ShareComboCommand{
		ComboID:   id,
		ShopperID: "owner-1",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	combo, _ := repo.FindById(context.Background(), id)
	assert.Equal(t, domain.VisibilityPublic, combo.Visibility())
	require.NotNil(t, combo.ShareToken())
	assert.Equal(t, token, *combo.ShareToken())
}

// ---------------------------------------------------------------------------
// 4.9 MakePrivateHandler
// ---------------------------------------------------------------------------

func TestMakePrivateHandler_AnyComboBecomePrivateWithNoToken(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		repo := newFakeRepo()
		id := seedCombo(t, repo, "owner-1", 3)
		// Share first so we start from a non-private state
		share := application.NewShareComboHandler(repo, services.NewShareTokenService(repo))
		_, err := share.Handle(context.Background(), application.ShareComboCommand{ComboID: id, ShopperID: "owner-1"})
		if err != nil {
			t.Fatalf("share: %v", err)
		}
		priv := application.NewMakePrivateHandler(repo)
		err = priv.Handle(context.Background(), application.MakePrivateCommand{ComboID: id, ShopperID: "owner-1"})
		if err != nil {
			t.Fatalf("make private: %v", err)
		}
		combo, _ := repo.FindById(context.Background(), id)
		if combo.Visibility() != domain.VisibilityPrivate {
			t.Fatalf("expected private, got %v", combo.Visibility())
		}
		if combo.ShareToken() != nil {
			t.Fatal("expected nil token after MakePrivate")
		}
	})
}

// ---------------------------------------------------------------------------
// 4.10 GetComboHandler — non-owner denied
// ---------------------------------------------------------------------------

func TestGetComboHandler_NonOwnerDenied(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-1", 2)
	h := application.NewGetComboHandler(repo, newStubEnrichment())
	_, err := h.Handle(context.Background(), application.GetComboQuery{
		ComboID:   id,
		ShopperID: "intruder",
	})
	assert.ErrorIs(t, err, domain.ErrComboAccessDenied)
}

// ---------------------------------------------------------------------------
// 4.11 GetSharedComboHandler
// ---------------------------------------------------------------------------

func TestGetSharedComboHandler_ValidTokenReturnsEnrichedCombo(t *testing.T) {
	repo := newFakeRepo()
	id := seedCombo(t, repo, "owner-1", 3)
	share := application.NewShareComboHandler(repo, services.NewShareTokenService(repo))
	token, err := share.Handle(context.Background(), application.ShareComboCommand{ComboID: id, ShopperID: "owner-1"})
	require.NoError(t, err)

	h := application.NewGetSharedComboHandler(repo, newStubEnrichment())
	combo, err := h.Handle(context.Background(), application.GetSharedComboQuery{ShareToken: token})
	require.NoError(t, err)
	assert.NotNil(t, combo)
}

func TestGetSharedComboHandler_InvalidTokenNotFound(t *testing.T) {
	repo := newFakeRepo()
	h := application.NewGetSharedComboHandler(repo, newStubEnrichment())
	_, err := h.Handle(context.Background(), application.GetSharedComboQuery{ShareToken: "ghost-token"})
	assert.ErrorIs(t, err, domain.ErrComboNotFound)
}
