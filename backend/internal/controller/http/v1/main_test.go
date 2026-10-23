package v1

import (
	"os"
	"testing"

	"github.com/TolyanL/beevibe/pkg/transaction"
)

func TestMain(m *testing.M) {
	transaction.IsUnitTest = true

	os.Exit(m.Run())
}
