package usecase

import (
	"context"
	"errors"
	"fmt"
)

// validateMaterialComposition checks the proposed materials before changing a
// product's composition. Reaching productID is always a cycle, regardless of its
// current materials, which may be replaced by the caller.
func validateMaterialComposition(ctx context.Context, repo MaterialRepository, productID int64, materialIDs []int64) error {
	if repo == nil {
		return errors.New("material repository is nil")
	}

	active := make(map[int64]bool)
	visited := make(map[int64]bool)
	var visit func(int64) error
	visit = func(id int64) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("request cancelled: %w", err)
		}
		if id == productID || active[id] {
			return fmt.Errorf("material composition contains a cycle at product %d", id)
		}
		if visited[id] {
			return nil
		}

		active[id] = true
		defer delete(active, id)
		materials, err := repo.GetMaterialsByProduct(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get materials for product %d: %w", id, err)
		}
		for _, material := range materials {
			materialID := material.MaterialID
			if materialID == 0 {
				materialID = material.ID
			}
			if err := visit(materialID); err != nil {
				return err
			}
		}
		visited[id] = true
		return nil
	}

	for _, id := range materialIDs {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
