package hills

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCoreDependsOnNothingOutside(t *testing.T) {
	core := []string{
		"github.com/omjogani/shape-hill/internal/hills",
		"github.com/omjogani/shape-hill/internal/account",
		"github.com/omjogani/shape-hill/internal/hillchart",
	}
	banned := []string{
		"net/http",
		"database/sql",
		"github.com/jackc/",
		"github.com/golang-jwt/",
		"github.com/MicahParks/",
		"github.com/spf13/",
		"github.com/omjogani/shape-hill/internal/store",
		"github.com/omjogani/shape-hill/internal/server",
		"github.com/omjogani/shape-hill/internal/config",
	}

	out, err := exec.Command("go", append([]string{"list", "-deps"}, core...)...).Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}

	for dep := range strings.FieldsSeq(string(out)) {
		for _, prefix := range banned {
			if strings.HasPrefix(dep, prefix) {
				t.Errorf("the core imports %s, which belongs outside it", dep)
			}
		}
	}
}
