package chainnotifymq

func init() {
	register(
		new(Redis),
		new(Kafka),
	)
}
