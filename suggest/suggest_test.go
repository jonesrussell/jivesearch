package suggest

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func createMockFile(f string) error {
	appFs = afero.NewMemMapFs()

	fh, err := appFs.OpenFile(f, os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("unable to append")
	}

	wrds := []string{"# a comment that should be skipped", "this is a bad phrase", "naughty", "really bad", "bad phrase"}

	for _, wrd := range wrds {
		if _, err := io.WriteString(fh, wrd+"\n"); err != nil {
			return err
		}
	}

	fh.Close()

	_, err = appFs.Stat(f)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %q does not exist", f)
	}

	return err
}

func TestNaughty(t *testing.T) {
	for _, c := range []struct {
		name string
		want bool
	}{
		{
			"this is a safe phrase",
			false,
		},
		{
			"naughty",
			true,
		},
		{
			"the bad bAD pHrase here",
			true,
		},
		{
			"really",
			false,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := "naughty.txt"
			if err := createMockFile(f); err != nil {
				t.Fatal(err)
			}

			if err := NewNaughty(f); err != nil {
				t.Fatal(err)
			}

			got := Naughty(c.name)

			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %t, want %t", got, c.want)
			}
		})
	}
}
