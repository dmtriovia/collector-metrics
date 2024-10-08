package storage

type Repository interface {
	Store(string, string) error
	Get(string) (string, error)
}
