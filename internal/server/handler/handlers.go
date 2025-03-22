package handler

import (
	db "task-manager/internal/database/sqlc"
	"task-manager/internal/server/token"

	"task-manager/config"

	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	db *db.Service

	tokenMaker token.Maker

	config *config.Config
}

func NewHandler(db *db.Service, config *config.Config, tokenMaker token.Maker) *Handler {
	return &Handler{
		db:         db,
		tokenMaker: tokenMaker,
		config:     config,
	}

}

// Helper function to wrap VersionID in pgtype.Text
func toPgTypeText(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  value != "",
	}
}

func toPgTypeInt(value int16) pgtype.Int2 {
	return pgtype.Int2{
		Int16: value,
		Valid: value != 0,
	}
}
