package main

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// makeSecrets is a util function to build a SecretList from []string secret names
func makeSecrets(secretNames []string) *v1.SecretList {
	var secrets []v1.Secret
	for _, secretName := range secretNames {
		objMeta := metav1.ObjectMeta{Name: secretName}
		secrets = append(secrets, v1.Secret{ObjectMeta: objMeta})
	}
	return &v1.SecretList{Items: secrets}
}

// TestGetSecretsLatestVersion makes sure that getSecrets returns the latest version of a release
func TestGetSecretsLatestVersion(t *testing.T) {
	secrets := makeSecrets([]string{"my-apache-3", "my-apache-4", "my-apache-9", "my-apache-1"})
	for _, secret := range getSecrets("my-apache", "", secrets) {
		if secret.Name != "my-apache-9" {
			t.Errorf("Incorrect")
		}
		break
	}
}

// TestGetSecretsSubStr makes sure that getSecrets returns a secret containing the string
func TestGetSecretsSubStr(t *testing.T) {
	secrets := makeSecrets([]string{"nginx-rp", "nginx-proxy", "nginx-lb", "nginx-static-assets"})
	for _, secret := range getSecrets("lb", "", secrets) {
		if secret.Name != "nginx-lb" {
			t.Errorf("Incorrect value")
		}
		break
	}
}

// TestGetSecretsList makes sure that getSecrets returns a list of secrets proposal
func TestGetSecretsProposals(t *testing.T) {
	secrets := []string{"qa-env-cert", "dev-env-cert", "mongo-dump", "prod-env-cert", "mongo-admin"}
	results := getSecrets("env-cert", "", makeSecrets(secrets))
	if len(results) != 3 {
		t.Errorf("Bad results length")
	}
}
