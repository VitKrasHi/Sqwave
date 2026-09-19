package game

import "Sqwave/internal/domain/input"

// InputSource — откуда app берёт ввод. Реализуется в infra.
type InputSource interface {
	Poll() input.PlayerInput
}
