package opsservice

import (
	"github.com/gravitational/gravity/lib/constants"
	"github.com/gravitational/gravity/lib/ops"
	"github.com/gravitational/gravity/lib/storage"

	"github.com/gravitational/rigging"
	"github.com/gravitational/trace"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// GetSMTPConfig returns the cluster SMTP configuration
func (o *Operator) GetSMTPConfig(key ops.SiteKey) (storage.SMTPConfig, error) {
	client, err := o.GetKubeClient()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	data, err := getSMTPConfig(client.Core().Secrets(metav1.NamespaceSystem))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	config, err := storage.UnmarshalSMTPConfig(data)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return config, nil
}

// UpdateSMTPConfig updates the cluster SMTP configuration
func (o *Operator) UpdateSMTPConfig(key ops.SiteKey, config storage.SMTPConfig) error {
	client, err := o.GetKubeClient()
	if err != nil {
		return trace.Wrap(err)
	}

	return updateSMTPConfig(client.Core().Secrets(metav1.NamespaceSystem), config)
}

// DeleteSMTPConfig deletes the cluster SMTP configuration
func (o *Operator) DeleteSMTPConfig(key ops.SiteKey) error {
	client, err := o.GetKubeClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = rigging.ConvertError(client.Core().Secrets(metav1.NamespaceSystem).Delete(constants.SMTPSecret, nil))
	if trace.IsNotFound(err) {
		return trace.NotFound("no SMTP configuration found")
	}
	return trace.Wrap(err)
}

func getSMTPConfig(client corev1.SecretInterface) ([]byte, error) {
	secret, err := client.Get(constants.SMTPSecret, metav1.GetOptions{})
	err = rigging.ConvertError(err)
	if err != nil {
		if trace.IsNotFound(err) {
			return nil, trace.NotFound("no SMTP configuration found")
		}
		return nil, trace.Wrap(err)
	}

	data, ok := secret.Data[constants.ResourceSpecKey]
	if !ok {
		return nil, trace.NotFound("no SMTP configuration found")
	}

	return data, nil
}

func updateSMTPConfig(client corev1.SecretInterface, config storage.SMTPConfig) error {
	bytes, err := storage.MarshalSMTPConfig(config)
	if err != nil {
		return trace.Wrap(err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.SMTPSecret,
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				// Update SMTP configuration for monitoring
				constants.MonitoringType: constants.MonitoringTypeSMTP,
			},
		},
		Data: map[string][]byte{
			constants.ResourceSpecKey: bytes,
		},
		Type: v1.SecretTypeOpaque,
	}

	_, err = client.Create(secret)
	err = rigging.ConvertError(err)
	if err == nil {
		return nil
	}

	if !trace.IsAlreadyExists(err) {
		return trace.Wrap(err)
	}

	_, err = client.Update(secret)
	return trace.Wrap(rigging.ConvertError(err))
}