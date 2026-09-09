package opelog

type Queue interface {
	Put(string) error
	Get() (string, error)
	Close() error
}
