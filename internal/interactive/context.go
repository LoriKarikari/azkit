package interactive

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/samber/lo"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func PickContext(ctx context.Context, contexts []domain.TenantContext) (domain.TenantContext, error) {
	if len(contexts) == 0 {
		return domain.TenantContext{}, app.ContextNotFound("")
	}
	selected := contexts[0]
	options := lo.Map(contexts, func(item domain.TenantContext, _ int) huh.Option[domain.TenantContext] {
		return huh.NewOption(fmt.Sprintf("%s — %s (%s)", item.Name, item.TenantID, item.Status), item)
	})
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[domain.TenantContext]().
			Title("Select a context").
			Options(options...).
			Value(&selected),
	))
	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return domain.TenantContext{}, ErrCanceled
		}
		return domain.TenantContext{}, err
	}
	return selected, nil
}
