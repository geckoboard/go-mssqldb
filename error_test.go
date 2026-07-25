package mssql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
)

func TestServerError(t *testing.T) {

	originalErr := Error{Message: "underlying error"}
	sererErr := ServerError{sqlError: originalErr}

	// Verify that error message is backwards compatible
	oldMessage := "SQL Server had internal error"
	if newMessage := sererErr.Error(); newMessage != oldMessage {
		t.Fatalf("ServerError returned incompatible error message. Got '%s', wanted '%s'", newMessage, oldMessage)
	}

	// Verify that the underlying error is preserved
	unwrappedErr := sererErr.Unwrap()
	if underlyingErr, ok := unwrappedErr.(Error); !ok || underlyingErr.Message != originalErr.Message {
		t.Fatalf("ServerError did not preserve wrapped error. Got '%+v', wanted '%+v'", unwrappedErr, originalErr)
	}
}

func TestRetryableError(t *testing.T) {

	originalErr := driver.ErrBadConn
	retryableErr := RetryableError{err: originalErr}

	// Verify that the error message matches the original error's
	origMessage := originalErr.Error()
	if wrappedMessage := retryableErr.Error(); wrappedMessage != origMessage {
		t.Fatalf("RetryableError returned incorrect error message. Got '%s', wanted '%s'", wrappedMessage, origMessage)
	}

	// Verify that the underlying error is preserved
	unwrappedErr := retryableErr.Unwrap()
	if unwrappedErr != originalErr {
		t.Fatalf("RetryableError did not preserve wrapped error. Got '%+v', wanted '%+v'", unwrappedErr, originalErr)
	}

	// Verify that underlying error is correctly recognized
	if !retryableErr.Is(driver.ErrBadConn) {
		t.Fatalf("RetryableError wrapping driver.ErrBadConn does not report it is a driver.ErrBadConn error")
	}

}

func TestStreamError(t *testing.T) {
	t.Run("returns the formatted message when no error is wrapped", func(t *testing.T) {
		err := streamErrorf("unexpected token %d", 42)

		if got, want := err.Error(), "Invalid TDS stream: unexpected token 42"; got != want {
			t.Errorf("got message %q, want %q", got, want)
		}

		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			t.Errorf("got wrapped error %q, want nil", unwrapped)
		}
	})

	t.Run("unwraps to the original error when the format wraps it", func(t *testing.T) {
		origErr := &net.OpError{Op: "read", Net: "tcp", Err: os.ErrDeadlineExceeded}
		err := streamErrorf("Reading PLP type failed: %w", origErr)

		if got, want := err.Error(), "Invalid TDS stream: Reading PLP type failed: read tcp: i/o timeout"; got != want {
			t.Errorf("got message %q, want %q", got, want)
		}

		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Errorf("expected error to unwrap to a net timeout but got %q", err)
		}
	})
}

func TestBadStreamPanic(t *testing.T) {

	errMsg := "test error XYZ"
	err := errors.New(errMsg)

	defer func() {
		r := recover()
		e, ok := r.(error)
		if !ok || !strings.HasSuffix(e.Error(), errMsg) {
			t.Fatalf("unexpected error recovered from panic: "+
				"got error = '%+v', wanted error to end with '%s'", e, errMsg)
		}

		if !errors.Is(e, err) {
			t.Errorf("expected recovered error to unwrap to the original error but got %q", e)
		}
	}()

	badStreamPanic(err)

	t.Fatalf("badStreamPanic did not panic as expected when passed %+v", err)
}

func TestBadStreamPanicf(t *testing.T) {

	errfmt := "the error is '%s'"
	errMsg := "test error XYZ"
	expectedMsg := fmt.Sprintf(errfmt, errMsg)

	defer func() {
		r := recover()
		if e, ok := r.(error); !ok || !strings.HasSuffix(e.Error(), expectedMsg) {
			t.Fatalf("unexpected error recovered from panic: "+
				"got error = '%+v', wanted error to end with '%s'", e, expectedMsg)
		}
	}()

	badStreamPanicf(errfmt, errMsg)

	t.Fatalf("badStreamPanicf did not panic as expected when passed %s", expectedMsg)
}
