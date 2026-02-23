package entities

import (
	"time"

	"github.com/google/uuid"
)

// Adicione suas entidades de domínio aqui
// Exemplo:
//
// type ExampleID = uuid.UUID
//
// type Example struct {
//     ID        ExampleID
//     Name      string
//     CreatedAt time.Time
//     UpdatedAt time.Time
// }

var _ = time.Time{}
var _ = uuid.UUID{}
