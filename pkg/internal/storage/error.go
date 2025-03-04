package storage

type StorageError struct {
	Status  int
	Message string
}

func (e StorageError) Error() string {
	return e.Message
}
