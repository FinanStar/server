package crypto

type passwordManager interface {
	Hash(password string) (string, error)
	Compare(password, hash string) (bool, error)
}

type generator interface {
	SecureId(length int) (string, error)
	RandomString(length int) string
}

type Crypto interface {
	PasswordManager() passwordManager
	Generator() generator
}

type crypto struct{}

func (self crypto) PasswordManager() passwordManager {
	return passwordManagerImpl{}
}

func (self crypto) Generator() generator {
	return generatorImpl{}
}

func Create() Crypto {
	return crypto{}
}
