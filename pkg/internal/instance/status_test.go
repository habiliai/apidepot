package instance_test

import (
	"crypto/tls"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func throwError() error {
	e := &tls.CertificateVerificationError{}
	return errors.Wrapf(e, "test error")
}

func throwOtherError() error {
	return errors.New("test error")
}

func TestErrorChecking(t *testing.T) {
	err := throwError()
	require.Error(t, err)
	if e := (*tls.CertificateVerificationError)(nil); !errors.As(err, &e) {
		require.FailNow(t, "error must be of type *tls.CertificateVerificationError")
	} else {
		require.NotNil(t, e)
		t.Logf("error is of type %T, error %+v", e, e)
	}

	err = throwOtherError()
	if e := (*tls.CertificateVerificationError)(nil); errors.As(err, &e) {
		require.FailNow(t, "error must be not of type *tls.CertificateVerificationError")
	} else {
		require.Nil(t, e)
		t.Logf("error is of type %T, error %+v", e, e)
	}
}
