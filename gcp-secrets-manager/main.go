package main

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"regexp"
)

type GcpSecretsManager struct {
	secretManagerClient    *secretmanager.Client
	CredentialsFilePath    string
	CredentialsFileContent string
}

type GcpCredentials struct {
	FilePath    string
	FileContent string
}

type GcpSecret struct {
	Name    string
	Value   string
	Project string
	Version string
}

func (m *GcpSecretsManager) WithCredentials(opts GcpCredentials) *GcpSecretsManager {
	m.CredentialsFileContent = opts.FilePath
	m.CredentialsFileContent = opts.FileContent

	return m
}

func (m *GcpSecretsManager) auth(ctx context.Context) error {
	var gcpOption option.ClientOption

	if m.CredentialsFilePath != "" {
		gcpOption = option.WithCredentialsFile(m.CredentialsFilePath)
	}

	if m.CredentialsFileContent != "" {
		gcpOption = option.WithCredentialsJSON([]byte(m.CredentialsFileContent))
	}

	client, err := secretmanager.NewClient(ctx, gcpOption)
	if err != nil {
		return err
	}

	m.secretManagerClient = client

	return nil
}

func (m *GcpSecretsManager) GetSecret(ctx context.Context, secret GcpSecret) (string, error) {
	err := m.auth(ctx)
	if err != nil {
		return "", err
	}

	if secret.Version == "" {
		secret.Version = "latest"
	}

	response, err := m.secretManagerClient.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", secret.Project, secret.Name, secret.Version),
	})
	if err != nil {
		return "", err
	}

	return string(response.Payload.GetData()), nil
}

func (m *GcpSecretsManager) SetSecret(ctx context.Context, secret GcpSecret) (string, error) {
	notFoundRegexp, err := regexp.Compile("rpc error: code = NotFound desc = Secret \\[projects/\\w+/secrets/\\w+] not found")
	if err != nil {
		return "", err
	}

	err = m.auth(ctx)
	if err != nil {
		return "", err
	}

	secretName := fmt.Sprintf("projects/%s/secrets/%s", secret.Project, secret.Name)
	_, err = m.secretManagerClient.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: secretName,
	})

	if err != nil {
		if !notFoundRegexp.MatchString(err.Error()) {
			return "", err
		}

		err = m.createSecret(ctx, secret)
		if err != nil {
			return "", err
		}
	}

	resp, err := m.secretManagerClient.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secret.Value),
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Name, nil
}

func (m *GcpSecretsManager) createSecret(ctx context.Context, secret GcpSecret) error {
	_, err := m.secretManagerClient.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
		SecretId: secret.Name,
		Parent:   fmt.Sprintf("projects/%s", secret.Project),
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	})

	return err
}
