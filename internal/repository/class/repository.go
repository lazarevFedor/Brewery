package class

import (
	"Brewery/internal/entities"
	"context"
	"Brewery/internal/repository/queries"


	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{
	Pool *pgxpool.Pool
	queries 
}

func NewRepository(pool *pgxpool.Pool) *Repository{
	return &Repository{Pool: pool}
}



func (repo *Repository) GetClasses(ctx context.Context) ([]entities.Class, error){
	
}

func (repo *Repository) InsertClass(ctx context.Context, tx pgx.Tx, class entities.Class) (uint, error)
func (repo *Repository) GetClassByID(ctx context.Context, id uint) (*entities.Class, error)
func (repo *Repository) UpdateClass(ctx context.Context, id uint, updates map[string]any) error
func (repo *Repository) DeleteClass(ctx context.Context, id uint) error
func (repo *Repository) GetClassID(ctx context.Context, tx pgx.Tx, ctgName string) (uint, error)
