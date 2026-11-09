package net

// WireInput — ввод игрока за кадр.
type WireInput struct {
	Up, Down, Left, Right                              bool
	Dash, Fire, FireReleased, SwapWeapon, Interact     bool
	MenuUp, MenuDown, MenuLeft, MenuRight, MenuConfirm bool
	DashChargeHeld                                     bool
	AimX, AimY                                         float64
}

// LobbyPlayer — игрок в лобби.
type LobbyPlayer struct {
	ID   uint8
	Name string
}

// EventKind — тип события из Poll.
type EventKind uint8

const (
	EvLobby        EventKind = iota // состав лобби изменился
	EvStart                         // хост начал игру
	EvSnapshot                      // снапшот мира
	EvGameOver                      // конец игры
	EvDisconnected                  // соединение потеряно/закрыто
	EvError                         // прочая ошибка (текст в Event.Err)
)

// Event — событие хоста или клиента.
type Event struct {
	Kind     EventKind
	Players  []LobbyPlayer // EvLobby
	Start    *StartInfo    // EvStart
	Snapshot *SnapshotMsg  // EvSnapshot
	Over     *GameOverMsg  // EvGameOver
	Err      string        // EvError / EvDisconnected
}

// StartInfo — параметры старта, рассылаемые хостом.
type StartInfo struct {
	Seed       int64
	WavesLimit int
	MapKind    int32
	Players    []LobbyPlayer
}

// SnapshotMsg — состояние мира, непрозрачное для пакета.
type SnapshotMsg struct {
	Tick uint32
	Data []byte
}

// GameOverMsg — результат матча, непрозрачные данные для пакета.
type GameOverMsg struct {
	Victory bool
	Data    []byte
}

// CommandKind — вид команды клиента.
type CommandKind uint8

const (
	// CmdSetLoadout — игрок выбрал снаряжение в меню спавна.
	CmdSetLoadout CommandKind = iota + 1
	// CmdBuyHP — игрок купил полное лечение.
	CmdBuyHP
	// CmdRevive — игрок купил воскрешение павших напарников.
	CmdRevive
)

// Command — команда клиента, применяемая хостом.
type Command struct {
	PlayerID  uint8 // проставляет хост при получении
	Kind      CommandKind
	Primary   int32 // WeaponType для CmdSetLoadout
	Secondary int32
}

// PlayerInput — ввод от клиента (для хоста).
type PlayerInput struct {
	PlayerID uint8
	Input    WireInput
}
