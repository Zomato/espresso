package certmanager

import "context"

type CredentialStore struct {
	credentials map[string]*SigningCredentials
}

func NewCredentialStore(credentials map[string]CertificateConfig) (*CredentialStore, error) {

	credentialStore := CredentialStore{
		credentials: make(map[string]*SigningCredentials),
	}

	for key, config := range credentials {
		cred, err := LoadSigningCredentials(context.Background(), &config)
		if err != nil {
			return nil, err
		}

		credentialStore.credentials[key] = cred
	}

	return &credentialStore, nil
}

func (cs *CredentialStore) GetCredential(key string) (*SigningCredentials, bool) {
	creds, exists := cs.credentials[key]
	return creds, exists
}
