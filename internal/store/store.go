package store

type Store struct {
	path string
}

func Open(path string) (*Store, error) {
	return &Store{path: path}, nil
}

func (s *Store) Set(key, value []byte) error {
	return nil
}

func (s *Store) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (s *Store) Delete(key []byte) error {
	return nil
}

func (s *Store) Close() error {
	return nil
}
