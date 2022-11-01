package storage

func init() {
	register(
		new(Std), new(PGSQL),
	)
}
