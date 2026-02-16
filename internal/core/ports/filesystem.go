package ports

type FilesystemPort interface {
	// SafeMove performs a cross-device aware move with retry logic.
	// 1. Try atomic rename
	// 2. If EXDEV, copy + verify hash + delete original
	SafeMove(src, dst string) error
}
