package storage

func init() {
	register(
		new(PGSQL),
	)
}
